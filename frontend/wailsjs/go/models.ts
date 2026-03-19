export namespace models {
	
	export class AIConfig {
	    provider: string;
	    model: string;
	    apiKey: string;
	    cloudflareAcct: string;
	    localEndpoint: string;
	
	    static createFrom(source: any = {}) {
	        return new AIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.apiKey = source["apiKey"];
	        this.cloudflareAcct = source["cloudflareAcct"];
	        this.localEndpoint = source["localEndpoint"];
	    }
	}
	export class Alert {
	    id: string;
	    type: string;
	    severity: string;
	    title: string;
	    message: string;
	    timestamp: number;
	    dismissed: boolean;
	    data?: any;
	
	    static createFrom(source: any = {}) {
	        return new Alert(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.severity = source["severity"];
	        this.title = source["title"];
	        this.message = source["message"];
	        this.timestamp = source["timestamp"];
	        this.dismissed = source["dismissed"];
	        this.data = source["data"];
	    }
	}
	export class AlertConfig {
	    cpuThreshold: number;
	    memoryThreshold: number;
	    diskThreshold: number;
	    enableAlerts: boolean;
	    enableSound: boolean;
	    enableDesktopNotf: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AlertConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpuThreshold = source["cpuThreshold"];
	        this.memoryThreshold = source["memoryThreshold"];
	        this.diskThreshold = source["diskThreshold"];
	        this.enableAlerts = source["enableAlerts"];
	        this.enableSound = source["enableSound"];
	        this.enableDesktopNotf = source["enableDesktopNotf"];
	    }
	}
	export class AutoInsight {
	    id: string;
	    title: string;
	    message: string;
	    category: string;
	    severity: string;
	    timestamp: number;
	    isRead: boolean;
	    data: string;
	    actionItems: string[];
	
	    static createFrom(source: any = {}) {
	        return new AutoInsight(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.message = source["message"];
	        this.category = source["category"];
	        this.severity = source["severity"];
	        this.timestamp = source["timestamp"];
	        this.isRead = source["isRead"];
	        this.data = source["data"];
	        this.actionItems = source["actionItems"];
	    }
	}
	export class ChatMessage {
	    id: string;
	    role: string;
	    content: string;
	    riskLevel?: string;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.riskLevel = source["riskLevel"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class ChatSession {
	    id: string;
	    title: string;
	    messages: ChatMessage[];
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ChatSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.messages = this.convertValues(source["messages"], ChatMessage);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ChatSessionSummary {
	    id: string;
	    title: string;
	    createdAt: number;
	    updatedAt: number;
	    messageCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ChatSessionSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.messageCount = source["messageCount"];
	    }
	}
	export class ConnectionInfo {
	    localAddr: string;
	    remoteAddr: string;
	    remoteHost: string;
	    country: string;
	    countryCode: string;
	    city: string;
	    latitude: number;
	    longitude: number;
	    processName: string;
	    pid: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.localAddr = source["localAddr"];
	        this.remoteAddr = source["remoteAddr"];
	        this.remoteHost = source["remoteHost"];
	        this.country = source["country"];
	        this.countryCode = source["countryCode"];
	        this.city = source["city"];
	        this.latitude = source["latitude"];
	        this.longitude = source["longitude"];
	        this.processName = source["processName"];
	        this.pid = source["pid"];
	    }
	}
	export class ContainerPort {
	    privatePort: number;
	    publicPort?: number;
	    type: string;
	    ip: string;
	
	    static createFrom(source: any = {}) {
	        return new ContainerPort(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.privatePort = source["privatePort"];
	        this.publicPort = source["publicPort"];
	        this.type = source["type"];
	        this.ip = source["ip"];
	    }
	}
	export class DevEnvironment {
	    id: string;
	    name: string;
	    type: string;
	    technology: string;
	    port: number;
	    processName: string;
	    processPID: number;
	    containerID?: string;
	    status: string;
	    icon: string;
	    description: string;
	    urls?: string[];
	
	    static createFrom(source: any = {}) {
	        return new DevEnvironment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.technology = source["technology"];
	        this.port = source["port"];
	        this.processName = source["processName"];
	        this.processPID = source["processPID"];
	        this.containerID = source["containerID"];
	        this.status = source["status"];
	        this.icon = source["icon"];
	        this.description = source["description"];
	        this.urls = source["urls"];
	    }
	}
	export class DevPort {
	    port: number;
	    processName: string;
	    processPID: number;
	    technology: string;
	    framework: string;
	    icon: string;
	    description: string;
	    url?: string;
	
	    static createFrom(source: any = {}) {
	        return new DevPort(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.processName = source["processName"];
	        this.processPID = source["processPID"];
	        this.technology = source["technology"];
	        this.framework = source["framework"];
	        this.icon = source["icon"];
	        this.description = source["description"];
	        this.url = source["url"];
	    }
	}
	export class DockerContainer {
	    id: string;
	    name: string;
	    image: string;
	    status: string;
	    state: string;
	    ports: ContainerPort[];
	    labels: Record<string, string>;
	    command: string;
	    createdAt: number;
	    startedAt: number;
	    finishedAt?: number;
	    exitCode?: number;
	    cpuPercent: number;
	    memoryMB: number;
	    networkRX: number;
	    networkTX: number;
	
	    static createFrom(source: any = {}) {
	        return new DockerContainer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.image = source["image"];
	        this.status = source["status"];
	        this.state = source["state"];
	        this.ports = this.convertValues(source["ports"], ContainerPort);
	        this.labels = source["labels"];
	        this.command = source["command"];
	        this.createdAt = source["createdAt"];
	        this.startedAt = source["startedAt"];
	        this.finishedAt = source["finishedAt"];
	        this.exitCode = source["exitCode"];
	        this.cpuPercent = source["cpuPercent"];
	        this.memoryMB = source["memoryMB"];
	        this.networkRX = source["networkRX"];
	        this.networkTX = source["networkTX"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DevEnvironmentInfo {
	    containers: DockerContainer[];
	    environments: DevEnvironment[];
	    devPorts: DevPort[];
	    dockerRunning: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DevEnvironmentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.containers = this.convertValues(source["containers"], DockerContainer);
	        this.environments = this.convertValues(source["environments"], DevEnvironment);
	        this.devPorts = this.convertValues(source["devPorts"], DevPort);
	        this.dockerRunning = source["dockerRunning"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class NetworkUsage {
	    pid: number;
	    processName: string;
	    bytesSent: number;
	    bytesRecv: number;
	    uploadSpeed: number;
	    downloadSpeed: number;
	
	    static createFrom(source: any = {}) {
	        return new NetworkUsage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.processName = source["processName"];
	        this.bytesSent = source["bytesSent"];
	        this.bytesRecv = source["bytesRecv"];
	        this.uploadSpeed = source["uploadSpeed"];
	        this.downloadSpeed = source["downloadSpeed"];
	    }
	}
	export class PortInfo {
	    port: number;
	    protocol: string;
	    state: string;
	    localAddr: string;
	    remoteAddr: string;
	    pid: number;
	    processName: string;
	
	    static createFrom(source: any = {}) {
	        return new PortInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.state = source["state"];
	        this.localAddr = source["localAddr"];
	        this.remoteAddr = source["remoteAddr"];
	        this.pid = source["pid"];
	        this.processName = source["processName"];
	    }
	}
	export class PrivacyConfig {
	    shareProcessNames: boolean;
	    shareProcessDetails: boolean;
	    shareNetworkPorts: boolean;
	    shareConnectionIPs: boolean;
	    shareConnectionGeo: boolean;
	    shareSecurityInfo: boolean;
	    shareSystemStats: boolean;
	    anonymizeProcesses: boolean;
	    anonymizeConnections: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PrivacyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shareProcessNames = source["shareProcessNames"];
	        this.shareProcessDetails = source["shareProcessDetails"];
	        this.shareNetworkPorts = source["shareNetworkPorts"];
	        this.shareConnectionIPs = source["shareConnectionIPs"];
	        this.shareConnectionGeo = source["shareConnectionGeo"];
	        this.shareSecurityInfo = source["shareSecurityInfo"];
	        this.shareSystemStats = source["shareSystemStats"];
	        this.anonymizeProcesses = source["anonymizeProcesses"];
	        this.anonymizeConnections = source["anonymizeConnections"];
	    }
	}
	export class ProcessInfo {
	    pid: number;
	    name: string;
	    commandLine: string;
	    cpuPercent: number;
	    memoryMB: number;
	    status: string;
	    username: string;
	    parentPid: number;
	    createTime: number;
	    numThreads: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.commandLine = source["commandLine"];
	        this.cpuPercent = source["cpuPercent"];
	        this.memoryMB = source["memoryMB"];
	        this.status = source["status"];
	        this.username = source["username"];
	        this.parentPid = source["parentPid"];
	        this.createTime = source["createTime"];
	        this.numThreads = source["numThreads"];
	    }
	}
	export class PromptTemplate {
	    id: string;
	    name: string;
	    description: string;
	    prompt: string;
	    category: string;
	    icon: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.prompt = source["prompt"];
	        this.category = source["category"];
	        this.icon = source["icon"];
	    }
	}
	export class ResourceTimelinePoint {
	    timestamp: number;
	    cpuPercent: number;
	    memoryPercent: number;
	    diskPercent: number;
	    netUploadSpeed: number;
	    netDownSpeed: number;
	
	    static createFrom(source: any = {}) {
	        return new ResourceTimelinePoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.cpuPercent = source["cpuPercent"];
	        this.memoryPercent = source["memoryPercent"];
	        this.diskPercent = source["diskPercent"];
	        this.netUploadSpeed = source["netUploadSpeed"];
	        this.netDownSpeed = source["netDownSpeed"];
	    }
	}
	export class SuspiciousProc {
	    pid: number;
	    name: string;
	    reasons: string[];
	    riskLevel: string;
	
	    static createFrom(source: any = {}) {
	        return new SuspiciousProc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.reasons = source["reasons"];
	        this.riskLevel = source["riskLevel"];
	    }
	}
	export class SecurityInfo {
	    firewallEnabled: boolean;
	    firewallStatus: string;
	    suspiciousProcs: SuspiciousProc[];
	    openPorts: number;
	    listeningPorts: number;
	    externalConns: number;
	    unknownConns: ConnectionInfo[];
	
	    static createFrom(source: any = {}) {
	        return new SecurityInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.firewallEnabled = source["firewallEnabled"];
	        this.firewallStatus = source["firewallStatus"];
	        this.suspiciousProcs = this.convertValues(source["suspiciousProcs"], SuspiciousProc);
	        this.openPorts = source["openPorts"];
	        this.listeningPorts = source["listeningPorts"];
	        this.externalConns = source["externalConns"];
	        this.unknownConns = this.convertValues(source["unknownConns"], ConnectionInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SystemContext {
	    processes: ProcessInfo[];
	    ports: PortInfo[];
	    network: NetworkUsage[];
	    // Go type: time
	    timestamp: any;
	    cpuUsage: number;
	    memUsage: number;
	    diskUsage: number;
	    diskUsedGB: number;
	    diskTotalGB: number;
	    securityInfo?: SecurityInfo;
	
	    static createFrom(source: any = {}) {
	        return new SystemContext(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.processes = this.convertValues(source["processes"], ProcessInfo);
	        this.ports = this.convertValues(source["ports"], PortInfo);
	        this.network = this.convertValues(source["network"], NetworkUsage);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.cpuUsage = source["cpuUsage"];
	        this.memUsage = source["memUsage"];
	        this.diskUsage = source["diskUsage"];
	        this.diskUsedGB = source["diskUsedGB"];
	        this.diskTotalGB = source["diskTotalGB"];
	        this.securityInfo = this.convertValues(source["securityInfo"], SecurityInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SystemStats {
	    cpuPercent: number;
	    cpuPerCore: number[];
	    memoryPercent: number;
	    memoryUsedGB: number;
	    memoryTotalGB: number;
	    swapPercent: number;
	    swapUsedGB: number;
	    swapTotalGB: number;
	    diskPercent: number;
	    diskUsedGB: number;
	    diskTotalGB: number;
	    diskReadSpeed: number;
	    diskWriteSpeed: number;
	    netUploadSpeed: number;
	    netDownSpeed: number;
	    uptime: number;
	    loadAvg1: number;
	    loadAvg5: number;
	    loadAvg15: number;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpuPercent = source["cpuPercent"];
	        this.cpuPerCore = source["cpuPerCore"];
	        this.memoryPercent = source["memoryPercent"];
	        this.memoryUsedGB = source["memoryUsedGB"];
	        this.memoryTotalGB = source["memoryTotalGB"];
	        this.swapPercent = source["swapPercent"];
	        this.swapUsedGB = source["swapUsedGB"];
	        this.swapTotalGB = source["swapTotalGB"];
	        this.diskPercent = source["diskPercent"];
	        this.diskUsedGB = source["diskUsedGB"];
	        this.diskTotalGB = source["diskTotalGB"];
	        this.diskReadSpeed = source["diskReadSpeed"];
	        this.diskWriteSpeed = source["diskWriteSpeed"];
	        this.netUploadSpeed = source["netUploadSpeed"];
	        this.netDownSpeed = source["netDownSpeed"];
	        this.uptime = source["uptime"];
	        this.loadAvg1 = source["loadAvg1"];
	        this.loadAvg5 = source["loadAvg5"];
	        this.loadAvg15 = source["loadAvg15"];
	        this.timestamp = source["timestamp"];
	    }
	}

}

export namespace services {
	
	export class ModelInfo {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class ProviderInfo {
	    id: string;
	    name: string;
	    models: ModelInfo[];
	    requiresApiKey: boolean;
	    requiresAcctId: boolean;
	    requiresEndpoint: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProviderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.models = this.convertValues(source["models"], ModelInfo);
	        this.requiresApiKey = source["requiresApiKey"];
	        this.requiresAcctId = source["requiresAcctId"];
	        this.requiresEndpoint = source["requiresEndpoint"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace version {
	
	export class Info {
	    version: string;
	    gitCommit: string;
	    gitTag: string;
	    buildDate: string;
	    buildUser: string;
	    goVersion: string;
	    platform: string;
	    arch: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.gitCommit = source["gitCommit"];
	        this.gitTag = source["gitTag"];
	        this.buildDate = source["buildDate"];
	        this.buildUser = source["buildUser"];
	        this.goVersion = source["goVersion"];
	        this.platform = source["platform"];
	        this.arch = source["arch"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

