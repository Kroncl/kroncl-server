package dadata

type FindPartyRequest struct {
	Query      string `json:"query"`
	BranchType string `json:"branch_type,omitempty"` // MAIN, BRANCH
	Type       string `json:"type,omitempty"`        // LEGAL, INDIVIDUAL
}

type PartySuggestion struct {
	Value             string `json:"value"`
	UnrestrictedValue string `json:"unrestricted_value"`
	Data              struct {
		INN           string          `json:"inn"`
		KPP           string          `json:"kpp"`
		OGRN          string          `json:"ogrn"`
		OGRNDate      int64           `json:"ogrn_date"`
		HID           string          `json:"hid"`
		Type          string          `json:"type"`
		Name          PartyName       `json:"name"`
		OPF           PartyOPF        `json:"opf"`
		Management    PartyManagement `json:"management"`
		Address       PartyAddress    `json:"address"`
		State         PartyState      `json:"state"`
		BranchType    string          `json:"branch_type"`
		BranchCount   int             `json:"branch_count"`
		EmployeeCount int             `json:"employee_count"`
		Okpo          string          `json:"okpo"`
		Oktmo         string          `json:"oktmo"`
		Okato         string          `json:"okato"`
	} `json:"data"`
}

type PartyName struct {
	FullWithOPF  string `json:"full_with_opf"`
	ShortWithOPF string `json:"short_with_opf"`
	Full         string `json:"full"`
	Short        string `json:"short"`
}

type PartyOPF struct {
	Code  string `json:"code"`
	Full  string `json:"full"`
	Short string `json:"short"`
}

type PartyManagement struct {
	Name      string `json:"name"`
	Post      string `json:"post"`
	StartDate int64  `json:"start_date"`
}

type PartyAddress struct {
	Value             string `json:"value"`
	UnrestrictedValue string `json:"unrestricted_value"`
}

type PartyState struct {
	Status           string `json:"status"`
	ActualityDate    int64  `json:"actuality_date"`
	RegistrationDate int64  `json:"registration_date"`
	LiquidationDate  *int64 `json:"liquidation_date"`
}

type FindPartyResponse struct {
	Suggestions []PartySuggestion `json:"suggestions"`
}

// CounterpartyPreview — то, что вернёт Kroncl фронту после заполнения из DaData
type CounterpartyPreview struct {
	Name    string `json:"name"`
	INN     string `json:"inn"`
	KPP     string `json:"kpp,omitempty"`
	OGRN    string `json:"ogrn"`
	Address string `json:"address"`
	Type    string `json:"type"` // organization, person
}
