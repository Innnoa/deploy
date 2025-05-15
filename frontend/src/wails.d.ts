interface ComputerInfo {
    name: string;
    seed: string;
    ip: string;
    oa: string;
  }

  interface Package {
    ID: string;
    AppName: string;
    AppType: string;
    Path: string;
    WinFile: string;
    UOSFile: string;
    KylinFile: string;
    Status: string;
  }

  
  interface App {
    GetComputerInfo(): Promise<ComputerInfo>;
    GetAllPackages(): Promise<Package[]>;
  }
  
  interface Main {
    App: App;
  }
  
  interface Go {
    main: Main;
  }
  
  interface Window {
    go: Go;
  }