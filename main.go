package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func main() {
	var err error

	// Full connection string is not checked in as is sensitive information
	db, err = sqlx.Connect("sqlserver", fullConnString)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("Connected!\n")

	err = ReadPatients()
	if err != nil {
		log.Fatal(err.Error())
	}

	os.Exit(0)
}

func ReadPatients() error {
	ctx := context.Background()

	// Check if database is alive.
	err := db.PingContext(ctx)
	// if err != nil {
	// 	return err
	// }

	// tsql := fmt.Sprintf("select Customer.id, DOB, Gender, FirstName, LastName, Postcode	from Customer inner join Customer_Address on Customer_Address.CustomerID = Customer.ID where customer.ID in (1,	100, 123, 331, 333,	777, 3864, 3911, 3914, 3915, 3916, 3917, 3920, 3921, 7888, 35369, 35394, 36095, 37209, 37229, 37248, 37457)")

	tsql := fmt.Sprintf("select distinct Customer.id, DOB, Gender, FirstName, LastName, Postcode from customer inner join Customer_Address on Customer_Address.CustomerID = Customer.ID where FirstName NOT LIKE '%%[^a-z-'']%%' and LastName NOT LIKE '%%[^a-z-'']%%'")

	// Execute query
	rows, err := db.QueryContext(ctx, tsql)
	if err != nil {
		return err
	}

	defer rows.Close()

	patients := []*patient{}

	for rows.Next() {
		p := &patient{}

		err := rows.Scan(&p.ID, &p.DOB, &p.Gender, &p.FirstName, &p.LastName, &p.Postcode)
		if err != nil {
			return err
		}

		if PatientExists(patients, p) {
			continue
		}

		patients = append(patients, p)
	}

	count := len(patients)

	foundPatients := []*patient{}

	for i, p := range patients {
		fmt.Println(i, " - ", count)
		request, err := http.NewRequest("GET", "https://pdrnhsinformationapi-sit.internal.pushsvcs.com/api/v1/PatientInformation/GetNhsNumber", nil)
		if err != nil {
			continue
		}

		q := request.URL.Query()
		q.Add("patientId", fmt.Sprint(p.ID))
		q.Add("dob", p.DOB)
		q.Add("gender", fmt.Sprint(p.Gender))
		q.Add("givenName", p.FirstName)
		q.Add("familyName", p.LastName)
		q.Add("postCode", p.Postcode)
		request.URL.RawQuery = q.Encode()

		// client := &http.Client{}
		// response, err := client.Do(request)
		req := request.URL.String()
		response, err := http.Get(req)

		if err != nil {
			return err
		}

		if response.StatusCode < 200 || response.StatusCode > 299 {
			continue
		}

		foundPatients = append(foundPatients, p)

		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		var result NhsNumberResponse
		if err := json.Unmarshal(responseData, &result); err != nil { // Parse []byte to go struct pointer
			fmt.Println("Can not unmarshal JSON")
		}

		p.NHSNumber = result.Result.NhsNumber

		request, err = http.NewRequest("GET", "https://pdrnhsinformationapi-sit.internal.pushsvcs.com/api/v1/PatientInformation/GetByNhsNumber", nil)
		if err != nil {
			return err
		}

		q = request.URL.Query()
		q.Add("patientId", fmt.Sprint(p.ID))
		q.Add("dob", p.DOB)
		q.Add("nhsNumber", p.NHSNumber)
		request.URL.RawQuery = q.Encode()

		req = request.URL.String()
		response, err = http.Get(req)

		if err != nil {
			return err
		}

		responseData, err = ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		var infoReponse NhsInfoResponse
		if err := json.Unmarshal(responseData, &infoReponse); err != nil { // Parse []byte to go struct pointer
			fmt.Println("Can not unmarshal JSON")
		}

		p.ODSCode = infoReponse.Result.Practice.OrganisationDataServiceCode
	}

	file, _ := json.MarshalIndent(foundPatients, "", " ")

	_ = ioutil.WriteFile("patients.json", file, 0644)

	return nil
}

func PatientExists(patients []*patient, p *patient) bool {
	for _, patient := range patients {
		if patient.FirstName == p.FirstName && patient.LastName == p.LastName {
			return true
		}
	}

	return false
}
