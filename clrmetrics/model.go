package clrmetrics

type Payload struct {
	Metrics         []Metric `json:"metrics"`
	Service         string   `json:"service"`
	ServiceInstance string   `json:"serviceInstance"`
}

type Metric struct {
	CPU struct {
		UsagePercent float64 `json:"usagePercent"`
	} `json:"cpu"`

	Exception struct {
		ExThrown                float64 `json:"exThrown"`
		ExThrownPerSec          float64 `json:"exThrownPerSec"`
		FiltersPerSec           float64 `json:"filtersPerSec"`
		FinallysPerSec          float64 `json:"finallysPerSec"`
		ThrowToCatchDepthPerSec float64 `json:"throwToCatchDepthPerSec"`
	} `json:"exception"`

	GC struct {
		BytesInAllHeaps     float64 `json:"bytesInAllHeaps"`
		Gen0CollectCount    float64 `json:"gen0CollectCount"`
		Gen0HeapSize        float64 `json:"gen0HeapSize"`
		Gen1CollectCount    float64 `json:"gen1CollectCount"`
		Gen1HeapSize        float64 `json:"gen1HeapSize"`
		Gen2CollectCount    float64 `json:"gen2CollectCount"`
		Gen2HeapSize        float64 `json:"gen2HeapSize"`
		HeapMemory          float64 `json:"heapMemory"`
		TotalCommittedBytes float64 `json:"totalCommittedBytes"`
		TotalReservedBytes  float64 `json:"totalReservedBytes"`
	} `json:"gc"`

	Interops struct {
		CcWs             float64 `json:"ccWs"`
		Marshalling      float64 `json:"marshalling"`
		Stubs            float64 `json:"stubs"`
		TlbExportsPerSec float64 `json:"tlbExportsPerSec"`
		TlbImportsPerSec float64 `json:"tlbImportsPerSec"`
	} `json:"interops"`

	JIT struct {
		CilBytesJitted      float64 `json:"cilBytesJitted"`
		IlBytesJittedPerSec float64 `json:"ilBytesJittedPerSec"`
		MethodsJitted       float64 `json:"methodsJitted"`
		TimeInJIT           float64 `json:"timeInJIT"`
	} `json:"jit"`

	Loading struct {
		BytesInLoaderHeap       float64 `json:"bytesInLoaderHeap"`
		PercentTimeLoading      float64 `json:"percentTimeLoading"`
		TotalAppDomains         float64 `json:"totalAppDomains"`
		TotalAppDomainsUnloaded float64 `json:"totalAppDomainsUnloaded"`
		TotalAssemblies         float64 `json:"totalAssemblies"`
		TotalClassesLoaded      float64 `json:"totalClassesLoaded"`
		TotalLoadFailures       float64 `json:"totalLoadFailures"`
	} `json:"loading"`

	Network struct {
		BytesReceived          float64 `json:"bytesReceived"`
		BytesSent              float64 `json:"bytesSent"`
		ConnectionsEstablished float64 `json:"connectionsEstablished"`
		DatagramsReceived      float64 `json:"datagramsReceived"`
		DatagramsSent          float64 `json:"datagramsSent"`
	} `json:"network"`

	Security struct {
		LinkTimeChecks        float64 `json:"linkTimeChecks"`
		TimeInRTChecks        float64 `json:"timeInRTChecks"`
		TimeSigAuthenticating float64 `json:"timeSigAuthenticating"`
		TotalRuntimeChecks    float64 `json:"totalRuntimeChecks"`
	} `json:"security"`

	Thread struct {
		AvailableCompletionPortThreads float64 `json:"availableCompletionPortThreads"`
		AvailableWorkerThreads         float64 `json:"availableWorkerThreads"`
		ContentionRate                 float64 `json:"contentionRate"`
		CurrentLogicalThreads          float64 `json:"currentLogicalThreads"`
		CurrentPhysicalThreads         float64 `json:"currentPhysicalThreads"`
		CurrentQueueLength             float64 `json:"currentQueueLength"`
		MaxCompletionPortThreads       float64 `json:"maxCompletionPortThreads"`
		MaxWorkerThreads               float64 `json:"maxWorkerThreads"`
		TotalContentions               float64 `json:"totalContentions"`
	} `json:"thread"`

	Time string `json:"time"`
}
