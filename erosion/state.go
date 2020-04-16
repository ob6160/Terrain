package erosion

type State struct {
	IsRaining                                                                                    bool
	WaterIncrementRate, GravitationalConstant, PipeCrossSectionalArea, EvaporationRate, TimeStep float32
	SedimentCarryCapacity, SoilSuspensionRate, SoilDepositionRate, MaximalErodeDepth             float32
}