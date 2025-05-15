import React, { useEffect, useState } from 'react';
import { Button, Typography ,Input ,Modal} from 'antd';
import { createStyles } from 'antd-style';
import computerLogo from '../../assets/images/computer-logo.png';
import {InitClient,GetPrinterModels} from "../../../wailsjs/go/deploy/Deploy";
import { ExclamationCircleFilled} from '@ant-design/icons';
import { useAppContext } from '../../context/AppContext';

const { Text } = Typography;

const useStyles = createStyles(({ css }) => ({
  welcomeContainer: css`
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100vh;
    width: 78%;
    padding: 20px;
    background-color: #F4F5F7;
  `,
  monitorImage: css`
    width: calc(600 * 100vh / 1024);
    height: calc(367 * 100vh / 1024);
    margin-bottom: 30px;
    display: flex;
    justify-content: center;
    align-items: center;
  `,
  computerLogo: css`
    max-width: 100%;
    max-height: 100%;
  `,
  title: css`
    text-align: center;
    margin-bottom: 5px !important;
    font-size: 28px;
    font-weight: 900;
    color: #222222;
  `,
  subtitle: css`
    text-align: center;
    margin-bottom: 10px !important;
    font-size: 28px;
    font-weight: 900;
    color: #222222;
  `,
  connectingText: css`
    color: #8181A5;
    margin-bottom: 20px;
  `,
  inputContainer: css`
    display: flex;
    margin-bottom: 20px;
    gap: 10px;
    margin-top : 20px;
    `,
  inputGroup: css`
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    
  `,
  inputLabel: css`
    font-size: 14px;
    font-weight: 700;
    margin-bottom: 8px;
    color: #222222;
  `,
  startButton: css`
    background-color: #0052cc;
    width: 150px;
    height: 40px;
    &:hover {
    background-color: #013FBF !important;
  }
  `
}));
interface WelcomeProps {
    onStartClick: () => void;
}
const Welcome: React.FC<WelcomeProps> = ({ onStartClick }) => {
    const { styles } = useStyles();
    const [packages, setPackages] = useState<any[]>([]);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [server, setServer] = useState<string>('');
    const [port, setPort] = useState<string>('');
    const [connected, setConnected] = useState<boolean>(false);
    const [printerModels, setPrinterModels] = useState<string[]>([]);

    // 使用上下文
    const appContext = useAppContext();

  // useEffect(() => {
  //   // 调用后端的 GetAllPackages 方法
  //   const fetchPackages = async () => {
  //     try {
  //       // 使用 window.go.main.App 访问后端导出的方法
  //       const allPackages = await window.go.main.App.GetAllPackages();
  //       setPackages(allPackages);
  //       console.log('获取所有包信息成功:', allPackages);
  //     } catch (error) {
  //       console.error('获取包信息失败:', error);
  //     } finally {
  //       setIsLoading(false);
  //     }
  //   };

  //   fetchPackages();
  // }, []);

  // 添加连接服务器的函数
  const connectToServer = async () => {
    // 检查 server 和 port 是否已输入
    if (!server || !port) {
      Modal.confirm({
        title: 'Information',
        icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
        content: 'Please input the server and port.',
        okText: 'Confirm',
        centered: true,
        
      });
      return;
    }
    setIsLoading(true);
    // 设置定时器，3秒后将 isLoading 设置为 true
    setTimeout(() => {
      setIsLoading(false);
      setConnected(true);
    }, 3000);
    // try {
    //   // 初始化客户端连接
    //   const info = await InitClient(server, port);

    //   // ... 其他操作
    // } catch (error) {
    //   console.error('连接服务器失败:', error);
    // } finally {
    //   setIsLoading(false);
    // }
    try {
      const models = await GetPrinterModels();
      console.log('获取打印机型号成功:', models);
      setPrinterModels(models);

      // 同时保存到上下文中
      appContext.setPrinterModels(models);
      appContext.setServer(server);
      appContext.setPort(port);

    } catch (error) {
      console.error('连接服务器失败:', error);
    } finally {
      // setIsLoading(false);
    }
  };

  // 处理按钮点击
  const handleButtonClick = () => {
    if (connected) {
      // 如果已连接，调用 onStartClick 进入下一步
      onStartClick();
    } else {
      // 否则连接服务器
      connectToServer();
    }
  };

  return (
    <div className={styles.welcomeContainer}>
      <div className={styles.monitorImage}>
        <img src={computerLogo} alt="Computer" className={styles.computerLogo} />
      </div>
      
      <span className={styles.title} >Deploy Tool for</span>
      <span className={styles.subtitle} >Windows</span>
      
      <div className={styles.inputContainer}>
        <div className={styles.inputGroup}>
          <span className={styles.inputLabel}>Server:</span>
          <Input 
            placeholder="Please input" 
            value={server}
            onChange={(e) => setServer(e.target.value)}
            style={{ width: '230px' }}
          />
        </div>
        <div className={styles.inputGroup}>
          <span className={styles.inputLabel}>Port:</span>
          <Input 
            placeholder="Please input" 
            value={port}
            onChange={(e) => setPort(e.target.value)}
            style={{ width: '230px' }}
          />
        </div>
      </div>
      {/* <Text className={styles.connectingText}>Connecting to the Servers, please wait...</Text> */}
      <Text className={styles.connectingText}>
        {isLoading ? "Connecting to the server, please wait..." : (connected ? `已连接，找到 ${packages.length} 个包` : "")}
      </Text>
    <Button 
        type="primary" 
        className={styles.startButton} 
        onClick={handleButtonClick}
        loading={isLoading}
        disabled={isLoading}
      >
        {connected ? "Start" : "Connect"}
      </Button>
    </div>
  );
};

export default Welcome;