package main

type patient struct {
	ID        int
	DOB       string
	Gender    int
	FirstName string
	LastName  string
	Postcode  string
	ODSCode   string
	NHSNumber string
}

type NhsNumberResponse struct {
	HTTPStatusCode string `json:"httpStatusCode"`
	Result         struct {
		NhsNumber string `json:"nhsNumber"`
	} `json:"result"`
}

type NhsInfoResponse struct {
	HTTPStatusCode string `json:"httpStatusCode"`
	Result         struct {
		Title      string   `json:"title"`
		NhsNumber  string   `json:"nhsNumber"`
		GivenNames []string `json:"givenNames"`
		FamilyName string   `json:"familyName"`
		Address    struct {
			AddressLines []string `json:"addressLines"`
			Postcode     string   `json:"postcode"`
			Use          string   `json:"use"`
		} `json:"address"`
		DateOfBirth string `json:"dateOfBirth"`
		Gender      struct {
			ID string `json:"id"`
		} `json:"gender"`
		Practice struct {
			OrganisationDataServiceCode string `json:"organisationDataServiceCode"`
			Name                        string `json:"name"`
			Address                     struct {
				AddressLines []string `json:"addressLines"`
				Postcode     string   `json:"postcode"`
			} `json:"address"`
			TelephoneNumber string `json:"telephoneNumber"`
		} `json:"practice"`
	} `json:"result"`
}
