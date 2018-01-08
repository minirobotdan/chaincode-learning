package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("mylogger")

// PersonalInfo schema
type PersonalInfo struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	DOB       string `json:"DOB"`
	Email     string `json:"email"`
	Mobile    string `json:"mobile"`
}

// FinancialInfo schema
type FinancialInfo struct {
	MonthlySalary      int `json:"monthlySalary"`
	MonthlyRent        int `json:"monthlyRent"`
	OtherExpenditure   int `json:"otherExpenditure"`
	MonthlyLoanPayment int `json:"monthlyLoanPayment"`
}

// LoanApplication schema
type LoanApplication struct {
	ID               string        `json:"id"`
	PropertyID       string        `json:"PropertyID"`
	LandID           string        `json:"LandID"`
	PermitID         string        `json:"PermitID"`
	BuyerID          string        `json:"BuyerID"`
	SalesContractID  string        `json:"SalesContractID"`
	PersonalInfo     PersonalInfo  `json:"personalInfo"`
	FinancialInfo    FinancialInfo `json:"financialInfo"`
	Status           string        `json:"status"`
	RequestedAmount  int           `json:"requestedAmount"`
	FairMarketValue  int           `json:"fairMarketValue"`
	ApprovedAmount   int           `json:"approvedAmount"`
	ReviewerID       string        `json:"ReviewerID"`
	LastModifiedDate string        `json:"lastModifiedDate"`
}

type customEvent struct {
	Type       string `json:"type"`
	Decription string `json:"description"`
}

// Sample chain code API
type SampleChainCode struct{}

// Stubbed init method
func (t *SampleChainCode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

// Query for existing
func (t *SampleChainCode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "GetLoanApplication" {
		return GetLoanApplication(stub, args)
	}
	return nil, nil
}

// Invoke creation of new application
func (t *SampleChainCode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "CreateLoanApplication" {
		username, _ := GetCertAttribute(stub, "username")
		role, _ := GetCertAttribute(stub, "role")
		if role == "Bank_Home_Loan_Admin" {
			return CreateLoanApplication(stub, args)
		}
		return nil, errors.Name(username + " with role " + role + " does not have correct permissions")
	}
}

func main() {

	lld, _ := shim.LogLevel("DEBUG")
	fmt.Println(lld)

	logger.SetLevel(lld)
	fmt.Println(logger.IsEnabledFor(lld))

	err := shim.Start(new(SampleChainCode))
	if err != nil {
		fmt.Println("Could not start SampleChainCode")
	} else {
		fmt.Println("SampleChainCode started successfully")
	}
}

// CreateLoanApplication Create loan application from args
func CreateLoanApplication(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Debug("Entering CreateLoanApplication")

	if len(args) < 2 {
		logger.Error("Invalid number of args")
		return nil, errors.New("Expected at least 2 arguments")
	}

	var loanAppID = args[0]
	var loanAppInput = args[1]

	err := stub.PutState(loanAppID, []byte(loanAppInput))
	if err != nil {
		logger.Error("Could not save loan application to ledger", err)
		return nil, err
	}

	var customEvent = "{eventType: 'loanApplicationCreation', description: '" + loanAppID + " Successfully created'}"
	err = stub.SetEvent("evtSender", []byte(customEvent))
	if err != nil {
		return nil, err
	}

	logger.Info("Successfully saved loan application")
	return nil, nil
}

// GetLoanApplication Get existing application by ID
func GetLoanApplication(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Debug("Entering GetLoanApplication")

	if len(args) < 1 {
		logger.Error("Invalid number of arguments")
		return nil, errors.New("Missing loan application ID")
	}

	var loanAppId = args[0]
	bytes, err := stub.GetState(loanAppId)
	if err != nil {
		logger.Error("Could not fetch loan application with id "+loanAppId+" from ledger", err)
		return nil, err
	}
	return bytes, nil
}

// UpdateLoanApplication Update existing application
func UpdateLoanApplication(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Debug("Entering UpdateLoanApplication")

	if len(args) < 2 {
		logger.Error("Invalid number of args")
		return nil, errors.New("Expected at least 2 arguments for loan application update")
	}

	var loanAppID = args[0]
	var status = args[1]

	laBytes, err := stub.GetState(loanAppId)
	if err != nil {
		logger.Error("Could not fetch loan application from ledger", err)
		return nil, err
	}
	var loanApplication loanApplication
	err = json.Unmarshal(laBytes, &loanApplication)
	loanApplication.Status = status

	laBytes, err = json.Marshal(&loanApplication)

	if err != nil {
		logger.Error("Could not marshal loan application post update", err)
		return nil, err
	}

	err = stub.PutState(loanAppID, laBytes)
	if err != nil {
		logger.Error("Could not save loan application post update", err)
		return nil, err
	}

	var customEvent = "{eventType: 'loanApplicationUpdate', description: '" + loanAppID + " Successfully updated'}"
	err = stub.SetEvent("evtSender", []byte(customEvent))
	if err != nil {
		return nil, err
	}

	logger.Info("Successfully updated loan application")
	return nil, nil
}

// GetCertAttribute Get particular attribute from JSON
func GetCertAttribute(stub shim.ChaincodeStubInterface, attributeName string) (string, error) {
	logger.Debug("Entering GetCertAttribute")
	attr, err := stub.ReadCertAttribute(attributeName)
	if err != nil {
		return "", errors.New("Couldn't get attribute " + attributeName + ". Error: " + err.Error())
	}
	attrString := string(attr)
	return attrString, nil
}
