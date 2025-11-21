export namespace common {
	
	export class ComputerInfo {
	    name: string;
	    seed: string;
	    ip: string;
	    oa: string;
	
	    static createFrom(source: any = {}) {
	        return new ComputerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.seed = source["seed"];
	        this.ip = source["ip"];
	        this.oa = source["oa"];
	    }
	}
	export class PackageInfo {
	    id: string;
	    appname: string;
	    brandId: string;
	    apptype: string;
	    installpath: string;
	    winfile: string;
	    uosdeb: string;
	    kylindeb: string;
	    status: string;
	    error: string;
	    pol: string;
	    ip: string;
	    reboot: string;
	    printerName: string;
	    printerDriver: string;
	    installPackageName: string;
	
	    static createFrom(source: any = {}) {
	        return new PackageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.appname = source["appname"];
	        this.brandId = source["brandId"];
	        this.apptype = source["apptype"];
	        this.installpath = source["installpath"];
	        this.winfile = source["winfile"];
	        this.uosdeb = source["uosdeb"];
	        this.kylindeb = source["kylindeb"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.pol = source["pol"];
	        this.ip = source["ip"];
	        this.reboot = source["reboot"];
	        this.printerName = source["printerName"];
	        this.printerDriver = source["printerDriver"];
	        this.installPackageName = source["installPackageName"];
	    }
	}
	export class Printer {
	    id: string;
	    pol: string;
	    ip: string;
	    appid: string;
	
	    static createFrom(source: any = {}) {
	        return new Printer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pol = source["pol"];
	        this.ip = source["ip"];
	        this.appid = source["appid"];
	    }
	}
	export class PrinterModel {
	    id: string;
	    brand: string;
	
	    static createFrom(source: any = {}) {
	        return new PrinterModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.brand = source["brand"];
	    }
	}
	export class SeedInfo {
	    seedlabel: string;
	    status: string;
	    errormsg: string;
	
	    static createFrom(source: any = {}) {
	        return new SeedInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seedlabel = source["seedlabel"];
	        this.status = source["status"];
	        this.errormsg = source["errormsg"];
	    }
	}

}

