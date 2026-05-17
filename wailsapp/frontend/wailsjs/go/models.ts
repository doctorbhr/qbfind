export namespace main {
	
	export class PreviewResult {
	    success: boolean;
	    text: string;
	    isText: boolean;
	    size: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new PreviewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.text = source["text"];
	        this.isText = source["isText"];
	        this.size = source["size"];
	        this.name = source["name"];
	    }
	}
	export class fileEntry {
	    path: string;
	    name: string;
	    lowerPath: string;
	    lowerName: string;
	    lowerExt: string;
	    size: number;
	    // Go type: time
	    modTime: any;
	    isDir: boolean;
	    priority: number;
	
	    static createFrom(source: any = {}) {
	        return new fileEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.lowerPath = source["lowerPath"];
	        this.lowerName = source["lowerName"];
	        this.lowerExt = source["lowerExt"];
	        this.size = source["size"];
	        this.modTime = this.convertValues(source["modTime"], null);
	        this.isDir = source["isDir"];
	        this.priority = source["priority"];
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

