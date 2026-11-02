package finance

const (
	// FilamentFailureBufferPercent defines the percentage of pure filament cost
	// allocated to the emergency fund (Kas Cadangan Risiko) to cover print failures (10%).
	FilamentFailureBufferPercent = 0.10

	// NetProfitEmergencyCutPercent defines the percentage deducted from positive net profit
	// to strengthen the emergency fund (Kas Cadangan Risiko) (5%).
	NetProfitEmergencyCutPercent = 0.05
)

