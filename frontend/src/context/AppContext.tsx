import React, { createContext, useState, useContext, ReactNode } from 'react';

// 定义上下文数据类型

interface PrinterModel {
  id: string;
  brand: string;
}
interface NetworkPinterModel {
  id: string;
  pol: string;
  ip: string;
  appid: string;
}
interface AppContextType {
  networkPinterModels: NetworkPinterModel[];
  printerModels: PrinterModel[];
  server: string;
  port: string;
  computerInfo: {
    name: string;
    seed: string;
    oa: string;
    ip: string;
  };
  setNetworkPinterModels: (models: NetworkPinterModel[]) => void;
  setPrinterModels: (models: PrinterModel[]) => void;
  setServer: (server: string) => void;
  setPort: (port: string) => void;
  setComputerInfo: (info: {name: string; seed: string; oa: string; ip: string}) => void;
}

// 创建上下文
const AppContext = createContext<AppContextType | undefined>(undefined);

// 创建上下文提供者组件
export const AppProvider: React.FC<{children: ReactNode}> = ({ children }) => {

  const [networkPinterModels, setNetworkPinterModels] = useState<NetworkPinterModel[]>([]); // 新增 NetworkPinterModels 状态和 setNetworkPinterModels 函数
  const [printerModels, setPrinterModels] = useState<PrinterModel[]>([]);
  const [server, setServer] = useState<string>('');
  const [port, setPort] = useState<string>('');
  const [computerInfo, setComputerInfo] = useState({
    name: '-',
    seed: '-',
    oa: '-',
    ip: '-.-.-.-'
  });

  // 提供上下文值
  const value = {
    networkPinterModels,
    printerModels,
    server,
    port,
    computerInfo,
    setNetworkPinterModels,
    setPrinterModels,
    setServer,
    setPort,
    setComputerInfo
  };

  return (
    <AppContext.Provider value={value}>
      {children}
    </AppContext.Provider>
  );
};

// 创建自定义钩子以便于使用上下文
export const useAppContext = () => {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useAppContext 必须在 AppProvider 内部使用');
  }
  return context;
};