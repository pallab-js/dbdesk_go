export namespace model {
	
	export class ActivityBucket {
	    start: number;
	    label: string;
	    success: number;
	    error: number;
	
	    static createFrom(source: any = {}) {
	        return new ActivityBucket(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start = source["start"];
	        this.label = source["label"];
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class Settings {
	    rowLimit: number;
	    timeoutSeconds: number;
	    historyEnabled: boolean;
	    retentionDays: number;
	    autoRefresh: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rowLimit = source["rowLimit"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.historyEnabled = source["historyEnabled"];
	        this.retentionDays = source["retentionDays"];
	        this.autoRefresh = source["autoRefresh"];
	    }
	}
	export class ConnectionInfo {
	    id: number;
	    name: string;
	    path: string;
	    readOnly: boolean;
	    lastOpenedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.readOnly = source["readOnly"];
	        this.lastOpenedAt = source["lastOpenedAt"];
	    }
	}
	export class AppState {
	    connection?: ConnectionInfo;
	    recent: ConnectionInfo[];
	    settings: Settings;
	    appDataDir: string;
	    demoPath: string;
	    version: string;
	    initialized: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connection = this.convertValues(source["connection"], ConnectionInfo);
	        this.recent = this.convertValues(source["recent"], ConnectionInfo);
	        this.settings = this.convertValues(source["settings"], Settings);
	        this.appDataDir = source["appDataDir"];
	        this.demoPath = source["demoPath"];
	        this.version = source["version"];
	        this.initialized = source["initialized"];
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
	export class Column {
	    name: string;
	    declType: string;
	    notNull: boolean;
	    default: string;
	    pk: number;
	    hidden: number;
	
	    static createFrom(source: any = {}) {
	        return new Column(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.declType = source["declType"];
	        this.notNull = source["notNull"];
	        this.default = source["default"];
	        this.pk = source["pk"];
	        this.hidden = source["hidden"];
	    }
	}
	export class ColumnMeta {
	    name: string;
	    type: string;
	    hidden: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ColumnMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.hidden = source["hidden"];
	    }
	}
	
	export class HistoryEntry {
	    id: number;
	    sql: string;
	    kind: string;
	    durationMs: number;
	    rowCount: number;
	    success: boolean;
	    error: string;
	    createdAt: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sql = source["sql"];
	        this.kind = source["kind"];
	        this.durationMs = source["durationMs"];
	        this.rowCount = source["rowCount"];
	        this.success = source["success"];
	        this.error = source["error"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class TelemetryTotals {
	    queries: number;
	    errors: number;
	    errorRate: number;
	    p50: number;
	    p95: number;
	    today: number;
	
	    static createFrom(source: any = {}) {
	        return new TelemetryTotals(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.queries = source["queries"];
	        this.errors = source["errors"];
	        this.errorRate = source["errorRate"];
	        this.p50 = source["p50"];
	        this.p95 = source["p95"];
	        this.today = source["today"];
	    }
	}
	export class LatencyPoint {
	    start: number;
	    label: string;
	    p50: number;
	    p95: number;
	
	    static createFrom(source: any = {}) {
	        return new LatencyPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start = source["start"];
	        this.label = source["label"];
	        this.p50 = source["p50"];
	        this.p95 = source["p95"];
	    }
	}
	export class TypeCount {
	    type: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new TypeCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.count = source["count"];
	    }
	}
	export class TableStat {
	    name: string;
	    rows: number;
	    approx: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new TableStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.rows = source["rows"];
	        this.approx = source["approx"];
	        this.error = source["error"];
	    }
	}
	export class ObjectCounts {
	    tables: number;
	    views: number;
	    indexes: number;
	    triggers: number;
	
	    static createFrom(source: any = {}) {
	        return new ObjectCounts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tables = source["tables"];
	        this.views = source["views"];
	        this.indexes = source["indexes"];
	        this.triggers = source["triggers"];
	    }
	}
	export class DatabaseStats {
	    path: string;
	    sizeBytes: number;
	    pageSize: number;
	    pageCount: number;
	    freePages: number;
	    journalMode: string;
	    encoding: string;
	    readOnly: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.sizeBytes = source["sizeBytes"];
	        this.pageSize = source["pageSize"];
	        this.pageCount = source["pageCount"];
	        this.freePages = source["freePages"];
	        this.journalMode = source["journalMode"];
	        this.encoding = source["encoding"];
	        this.readOnly = source["readOnly"];
	    }
	}
	export class DashboardData {
	    generatedAt: number;
	    database: DatabaseStats;
	    objects: ObjectCounts;
	    tables: TableStat[];
	    columnTypes: TypeCount[];
	    activity: ActivityBucket[];
	    latency: LatencyPoint[];
	    totals: TelemetryTotals;
	    recent: HistoryEntry[];
	    partial: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generatedAt = source["generatedAt"];
	        this.database = this.convertValues(source["database"], DatabaseStats);
	        this.objects = this.convertValues(source["objects"], ObjectCounts);
	        this.tables = this.convertValues(source["tables"], TableStat);
	        this.columnTypes = this.convertValues(source["columnTypes"], TypeCount);
	        this.activity = this.convertValues(source["activity"], ActivityBucket);
	        this.latency = this.convertValues(source["latency"], LatencyPoint);
	        this.totals = this.convertValues(source["totals"], TelemetryTotals);
	        this.recent = this.convertValues(source["recent"], HistoryEntry);
	        this.partial = source["partial"];
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
	
	export class ForeignKey {
	    from: string;
	    table: string;
	    to: string;
	    onUpdate: string;
	    onDelete: string;
	
	    static createFrom(source: any = {}) {
	        return new ForeignKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.table = source["table"];
	        this.to = source["to"];
	        this.onUpdate = source["onUpdate"];
	        this.onDelete = source["onDelete"];
	    }
	}
	
	export class IndexInfo {
	    name: string;
	    unique: boolean;
	    columns: string[];
	    partial: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IndexInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.unique = source["unique"];
	        this.columns = source["columns"];
	        this.partial = source["partial"];
	    }
	}
	export class KeyValue {
	    name: string;
	    value: any;
	
	    static createFrom(source: any = {}) {
	        return new KeyValue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	
	export class MutationResult {
	    affected: number;
	    lastInsertId: number;
	
	    static createFrom(source: any = {}) {
	        return new MutationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.affected = source["affected"];
	        this.lastInsertId = source["lastInsertId"];
	    }
	}
	
	export class QueryResult {
	    columns: ColumnMeta[];
	    rows: any[][];
	    rowCount: number;
	    truncated: boolean;
	    durationMs: number;
	    affected: number;
	    kind: string;
	    statement: string;
	    isRead: boolean;
	    lastInsertId: number;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = this.convertValues(source["columns"], ColumnMeta);
	        this.rows = source["rows"];
	        this.rowCount = source["rowCount"];
	        this.truncated = source["truncated"];
	        this.durationMs = source["durationMs"];
	        this.affected = source["affected"];
	        this.kind = source["kind"];
	        this.statement = source["statement"];
	        this.isRead = source["isRead"];
	        this.lastInsertId = source["lastInsertId"];
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
	export class RowDeleteRequest {
	    table: string;
	    keys: KeyValue[];
	
	    static createFrom(source: any = {}) {
	        return new RowDeleteRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = source["table"];
	        this.keys = this.convertValues(source["keys"], KeyValue);
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
	export class RowInsertRequest {
	    table: string;
	    values: KeyValue[];
	
	    static createFrom(source: any = {}) {
	        return new RowInsertRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = source["table"];
	        this.values = this.convertValues(source["values"], KeyValue);
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
	export class RowUpdateRequest {
	    table: string;
	    keys: KeyValue[];
	    values: KeyValue[];
	
	    static createFrom(source: any = {}) {
	        return new RowUpdateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = source["table"];
	        this.keys = this.convertValues(source["keys"], KeyValue);
	        this.values = this.convertValues(source["values"], KeyValue);
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
	export class SavedQuery {
	    id: number;
	    name: string;
	    sql: string;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new SavedQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sql = source["sql"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class Table {
	    name: string;
	    type: string;
	    columns: Column[];
	    fks: ForeignKey[];
	    indexes: IndexInfo[];
	    ddl: string;
	    rowKey: string;
	    editable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Table(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.columns = this.convertValues(source["columns"], Column);
	        this.fks = this.convertValues(source["fks"], ForeignKey);
	        this.indexes = this.convertValues(source["indexes"], IndexInfo);
	        this.ddl = source["ddl"];
	        this.rowKey = source["rowKey"];
	        this.editable = source["editable"];
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
	export class Schema {
	    tables: Table[];
	
	    static createFrom(source: any = {}) {
	        return new Schema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tables = this.convertValues(source["tables"], Table);
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
	
	
	export class TablePage {
	    table: string;
	    columns: ColumnMeta[];
	    rows: any[][];
	    keyIndexes: number[];
	    keyNames: string[];
	    total: number;
	    countError: string;
	    page: number;
	    pageSize: number;
	    hasNext: boolean;
	    editable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TablePage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = source["table"];
	        this.columns = this.convertValues(source["columns"], ColumnMeta);
	        this.rows = source["rows"];
	        this.keyIndexes = source["keyIndexes"];
	        this.keyNames = source["keyNames"];
	        this.total = source["total"];
	        this.countError = source["countError"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.hasNext = source["hasNext"];
	        this.editable = source["editable"];
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
	export class TablePageRequest {
	    table: string;
	    page: number;
	    pageSize: number;
	    sortColumn: string;
	    sortDir: string;
	
	    static createFrom(source: any = {}) {
	        return new TablePageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = source["table"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.sortColumn = source["sortColumn"];
	        this.sortDir = source["sortDir"];
	    }
	}
	
	

}

