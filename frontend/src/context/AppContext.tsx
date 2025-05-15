import React, { createContext, useState, useContext, ReactNode } from 'react';

// 定义上下文数据类型
interface AppContextType {
  printerModels: string[];
  server: string;
  port: string;
  setPrinterModels: (models: string[]) => void;
  setServer: (server: string) => void;
  setPort: (port: string) => void;
}

// 创建上下文
const AppContext = createContext<AppContextType | undefined>(undefined);

// 创建上下文提供者组件
export const AppProvider: React.FC<{children: ReactNode}> = ({ children }) => {
  const [printerModels, setPrinterModels] = useState<string[]>([]);
  const [server, setServer] = useState<string>('');
  const [port, setPort] = useState<string>('');

  // 提供上下文值
  const value = {
    printerModels,
    server,
    port,
    setPrinterModels,
    setServer,
    setPort
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