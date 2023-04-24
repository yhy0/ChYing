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

