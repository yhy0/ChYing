export namespace burpSuite {
	
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
	export class Setting {
	    port: number;
	    exclude: string[];
	    include: string[];
	    filterSuffix: string[];
	
	    static createFrom(source: any = {}) {
	        return new Setting(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.exclude = source["exclude"];
	        this.include = source["include"];
	        this.filterSuffix = source["filterSuffix"];
	    }
	}
	export class SettingUI {
	    port: number;
	    exclude: string;
	    include: string;
	    filterSuffix: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingUI(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.exclude = source["exclude"];
	        this.include = source["include"];
	        this.filterSuffix = source["filterSuffix"];
	    }
	}

}

export namespace main {
	
	export class Message {
	    msg: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.msg = source["msg"];
	        this.error = source["error"];
	    }
	}

}

export namespace nucleiY {
	
	export class Options {
	    label: string;
	    children: string[];
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.children = source["children"];
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

