export namespace data {
	
	export class HTTPBody {
	    targetUrl: string;
	    request: string;
	    response: string;
	    uuid: string;
	
	    static createFrom(source: any = {}) {
	        return new HTTPBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.targetUrl = source["targetUrl"];
	        this.request = source["request"];
	        this.response = source["response"];
	        this.uuid = source["uuid"];
	    }
	}

}

export namespace twj {
	
	export class Jwt {
	    header: string;
	    payload: string;
	    message: string;
	    signature: string;
	
	    static createFrom(source: any = {}) {
	        return new Jwt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.header = source["header"];
	        this.payload = source["payload"];
	        this.message = source["message"];
	        this.signature = source["signature"];
	    }
	}

}

