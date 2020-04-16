package meta

// #####################importOpts########################
type ImportSettings struct {
	vers string
	size uint64
	pwd  string
}

type ImportOption func(*ImportSettings) error

func ImportOptions(opts ...ImportOption) (*ImportSettings, error) {
	options := DefaultImportOptions()

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func DefaultImportOptions() *ImportSettings {
	return &ImportSettings{
		vers: "lotus",
		size: 1 << 20,
		pwd:  "",
	}
}

func ImportAPI(v string) ImportOption {
	return func(options *ImportSettings) error {
		options.vers = v
		return nil
	}
}

func SecSize(s uint64) ImportOption {
	return func(options *ImportSettings) error {
		options.size = s
		return nil
	}
}

func Pwd(p string) ImportOption {
	return func(options *ImportSettings) error {
		options.pwd = p
		return nil
	}
}

// #####################dealOpts########################
type DealSettings struct {
	vers  string
	askId uint64
	price string
}

type DealOption func(*DealSettings) error

func DealOptions(opts ...DealOption) (*DealSettings, error) {
	options := DefaultDealOptions()

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func DefaultDealOptions() *DealSettings {
	return &DealSettings{
		vers:  "lotus",
		askId: 0,
		price: "",
	}
}

func DealAPI(v string) DealOption {
	return func(options *DealSettings) error {
		options.vers = v
		return nil
	}
}

func AskId(a uint64) DealOption {
	return func(options *DealSettings) error {
		options.askId = a
		return nil
	}
}

func Price(p string) DealOption {
	return func(options *DealSettings) error {
		options.price = p
		return nil
	}
}

// #####################dealStateOpts########################
type DealStateSettings struct {
	vers string
}

type DealStateOption func(*DealStateSettings) error

func DealStateOptions(opts ...DealStateOption) (*DealStateSettings, error) {
	options := DefaultDealStateOptions()

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func DefaultDealStateOptions() *DealStateSettings {
	return &DealStateSettings{
		vers: "lotus",
	}
}

func DealStateAPI(v string) DealStateOption {
	return func(options *DealStateSettings) error {
		options.vers = v
		return nil
	}
}

// #####################retriveOpts########################
type RetrieveSettings struct {
	vers  string
	miner string
}

type RetrieveOption func(*RetrieveSettings) error

func RetrieveOptions(opts ...RetrieveOption) (*RetrieveSettings, error) {
	options := DefaultRetrieveOptions()

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func DefaultRetrieveOptions() *RetrieveSettings {
	return &RetrieveSettings{
		vers:  "lotus",
		miner: "",
	}
}

func RetrieveAPI(v string) RetrieveOption {
	return func(options *RetrieveSettings) error {
		options.vers = v
		return nil
	}
}

func RetrieveMiner(m string) RetrieveOption {
	return func(options *RetrieveSettings) error {
		options.miner = m
		return nil
	}
}
