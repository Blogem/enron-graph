export namespace explorer {
	
	export class PropertyDefinition {
	    name: string;
	    data_type: string;
	    sample_value?: string[];
	    nullable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PropertyDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.data_type = source["data_type"];
	        this.sample_value = source["sample_value"];
	        this.nullable = source["nullable"];
	    }
	}
	export class SchemaType {
	    name: string;
	    category: string;
	    count: number;
	    properties: PropertyDefinition[];
	    is_promoted: boolean;
	    relationships?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SchemaType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.category = source["category"];
	        this.count = source["count"];
	        this.properties = this.convertValues(source["properties"], PropertyDefinition);
	        this.is_promoted = source["is_promoted"];
	        this.relationships = source["relationships"];
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
	export class SchemaResponse {
	    promoted_types: SchemaType[];
	    discovered_types: SchemaType[];
	    total_entities: number;
	
	    static createFrom(source: any = {}) {
	        return new SchemaResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.promoted_types = this.convertValues(source["promoted_types"], SchemaType);
	        this.discovered_types = this.convertValues(source["discovered_types"], SchemaType);
	        this.total_entities = source["total_entities"];
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

