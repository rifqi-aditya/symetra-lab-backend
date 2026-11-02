package order

const (
	// FilamentFailureBufferPercent defines the percentage of pure filament cost
	// allocated to the emergency fund (Kas Cadangan Risiko) to cover print failures (10%).
	FilamentFailureBufferPercent = 0.10

	// GrossFilamentToBufferRatio is the derived ratio to extract the 10% emergency buffer
	// from a gross filament amount that already includes the 10% buffer: (0.10 / 1.10 ≈ 0.090909).
	GrossFilamentToBufferRatio = FilamentFailureBufferPercent / (1.0 + FilamentFailureBufferPercent)

	// NetProfitEmergencyCutPercent defines the percentage deducted from positive net profit
	// to strengthen the emergency fund (Kas Cadangan Risiko) (5%).
	NetProfitEmergencyCutPercent = 0.05
)

