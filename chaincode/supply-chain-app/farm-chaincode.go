// SPDX-License-Identifier: Apache-2.0

package main

/* Imports  
* 4 utility libraries for handling bytes, reading and writing JSON, 
formatting, and string manipulation  
* 2 specific Hyperledger Fabric specific libraries for Smart Contracts  
*/ 
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

/* Define Farm produce structure, with 6 properties.  
Structure tags are used by encoding/json library
*/
type Produce struct {
	Product string `json:"product"`  //selected from a predefined list of different product types
	Weight string `json:"weight"`    //received as float64 value
	Organic string `json:"organic"`  //received as boolean True or False
	Location  string `json:"location"` //received as longitude and latitude
	Timestamp string `json:"timestamp"` //received as date-time
	Holder  string `json:"holder"`
}

/*
 * The Init method *
 called when the Smart Contract "farm-chaincode" is instantiated by the network
 * Best practice is to have any Ledger initialization in separate function 
 -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method *
 called when an application requests to run the Smart Contract "farm-chaincode"
 The app also specifies the specific smart contract function to call with args
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger
	if function == "queryProduce" {
		return s.queryProduce(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "recordProduce" {
		return s.recordProduce(APIstub, args)
	} else if function == "queryAllProduce" {
		return s.queryAllProduce(APIstub)
	} else if function == "changeProduceHolder" {
		return s.changeProduceHolder(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

/*
 * The queryProduce method *
Used to view the records of one particular produce
It takes one argument -- the key for the produce  in question
 */
func (s *SmartContract) queryProduce(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ProduceAsBytes, _ := APIstub.GetState(args[0])
	if ProduceAsBytes == nil {
		return shim.Error("Could not locate produce, verify if the key number is correct!")
	}
	return shim.Success(ProduceAsBytes)
}

/*
 * The initLedger method *
Pre-populate the ledger with some initial data of farm produce
 */
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	produce := []Produce{
		Produce{Product: "Chicken", Weight: "1400.00", Organic: "true", Location: "67.0006, -70.5476", Timestamp: "Fri Jun 22 2018 11:02:01 GMT+0530 (India Standard Time)", Holder: "Sakya"},
		Produce{Product: "Beef", Weight: "1000.00", Organic: "false", Location: "91.2395, -49.4594", Timestamp: "Fri Jan 11 2019 12:01:01 GMT+0800 (Singapore Standard Time)", Holder: "Ilya"},
		Produce{Product: "Pork", Weight: "1200.00", Organic: "false", Location: "58.0148, 59.01391", Timestamp: "Fri Jan 11 2019 12:05:21 GMT+0800 (Singapore Standard Time)", Holder: "Dan"},
		Produce{Product: "Salmon", Weight: "1500.00", Organic: "true", Location: "-45.0945, 0.7949", Timestamp: "Wed Mar 13 2019 10:05:01 GMT+0800 (Singapore Standard Time)", Holder: "George"},
		Produce{Product: "Salmon", Weight: "2400.00", Organic: "true", Location: "-107.6043, 19.5003", Timestamp: "Fri Mar 15 2019 20:00:01 GMT+0800 (Singapore Standard Time)", Holder: "John"},
	}

	count := 0
	for count < len(produce) {
		fmt.Println("count is ", count)
		ProduceAsBytes, _ := json.Marshal(produce[count])
		APIstub.PutState(strconv.Itoa(count+1), ProduceAsBytes)
		fmt.Println("Added to the ledger", produce[count])
		count += 1
	}

	return shim.Success(nil)
}

/*
 * The recordProduce method *
This method takes in 7 arguments (6 produce attributes and Item number to be saved in the ledger). 
 */
func (s *SmartContract) recordProduce(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}
	
	// check if the first argument (key) is an integer
	if _, err := strconv.ParseInt(args[0],10,64); err != nil {
		return shim.Error(fmt.Sprintf("Failed to record into ledger for non integer ID: %s", args[0]))
	}
	
	// check if UID already exists in the ledger
    produceAsBytes, _ := APIstub.GetState(args[0])
	if produceAsBytes != nil {
		return shim.Error("Produce with the same ID already exists in the ledger, you may want to use changeProduceHolder function")
	}

	var produce = Produce{ Product: args[1], Weight: args[2], Organic: args[3], Location: args[4], Timestamp: args[5], Holder: args[6] }

	farmAsBytes, _ := json.Marshal(produce)
	err := APIstub.PutState(args[0], farmAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record into ledger for produce ID: %s", args[0]))
	}

	return shim.Success(nil)
}

/*
 * The queryAllProduce method *
Get the current global state of the ledger
This method does not take any arguments. Returns JSON string containing results. 
 */
func (s *SmartContract) queryAllProduce(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "0"
	endKey := "999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add comma before array members,suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- Global state of the ledger:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

/*
 * The changeProduceHolder method *
The data in the world state can be updated with who has current possession of produce. 
This function takes in 2 arguments, unique id and new holder name. 
 */
func (s *SmartContract) changeProduceHolder(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	produceAsBytes, _ := APIstub.GetState(args[0])
	if produceAsBytes == nil {
		return shim.Error("Could not locate produce record on ledger")
	}

	produce := Produce{}

	json.Unmarshal(produceAsBytes, &produce)

	//check validity of unmarsheled return
	//if produce.Holder == args[1] {
	//	return shim.Error(fmt.Sprintf("Illegal request for produce holder %s", args[1]))
	//}

	produce.Holder = args[1]

	produceAsBytes, _ = json.Marshal(produce)
	err := APIstub.PutState(args[0], produceAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to change produce holder: %s", args[0]))
	}

	return shim.Success(nil)
}

/*
 * main function *
calls the Start function 
The main function starts the chaincode in the container during instantiation.
 */
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}