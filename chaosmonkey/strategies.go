package chaosmonkey

// List of default chaos strategies supported by Chaos Monkey
const (
	// StrategyShutdownInstance ...
	StrategyShutdownInstance = "ShutdownInstance"

	// StrategyBlockAllNetworkTraffic ...
	StrategyBlockAllNetworkTraffic = "BlockAllNetworkTraffic"

	// StrategyDetachVolumes ...
	StrategyDetachVolumes = "DetachVolumes"

	// StrategyBurnCPU ...
	StrategyBurnCPU = "BurnCpu"

	// StrategyBurnIO ...
	StrategyBurnIO = "BurnIo"

	// StrategyKillProcesses ...
	StrategyKillProcesses = "KillProcesses"

	// StrategyNullRoute ...
	StrategyNullRoute = "NullRoute"

	// StrategyFailEC2 ...
	StrategyFailEC2 = "FailEc2"

	// StrategyFailDNS ...
	StrategyFailDNS = "FailDns"

	// StrategyFailDynamoDB ...
	StrategyFailDynamoDB = "FailDynamoDb"

	// StrategyFailS3 ...
	StrategyFailS3 = "FailS3"

	// StrategyFillDisk ...
	StrategyFillDisk = "FillDisk"

	// StrategyNetworkCorruption ...
	StrategyNetworkCorruption = "NetworkCorruption"

	// StrategyNetworkLatency ...
	StrategyNetworkLatency = "NetworkLatency"

	// StrategyNetworkLoss ...
	StrategyNetworkLoss = "NetworkLoss"
)