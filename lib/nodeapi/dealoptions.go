package api

type DealSettings struct {
	AskID      uint64
	Price      string
	WalletAddr string
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
		AskID:      0,
		Price:      "1",
		WalletAddr: "",
	}
}

func AskID(askid uint64) DealOption {
	return func(options *DealSettings) error {
		options.AskID = askid
		return nil
	}
}

func Price(price string) DealOption {
	return func(options *DealSettings) error {
		options.Price = price
		return nil
	}
}

func WalletAddr(addr string) DealOption {
	return func(options *DealSettings) error {
		options.WalletAddr = addr
		return nil
	}
}
