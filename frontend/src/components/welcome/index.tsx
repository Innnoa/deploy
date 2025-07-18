import React, { useState } from 'react';
import { Button, Typography ,Input ,Modal} from 'antd';
import { createStyles } from 'antd-style';
import computerLogo from '../../assets/images/computer-logo.png';
import {GetOAServer,GetSeedLabel,
        GetPrinterModels, GetNetworkPinterList, CheckSeedLabel
} from "../../../wailsjs/go/deploy/Deploy";
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
    height : 20px;
  `,
  // inputContainer: css`
  //   display: flex;
  //   margin-bottom: 20px;
  //   gap: 10px;
  //   margin-top : 20px;
  //   `,
  // inputGroup: css`
  //   flex: 1;
  //   display: flex;
  //   flex-direction: column;
  //   align-items: flex-start;
    
  // `,
  // inputLabel: css`
  //   font-size: 14px;
  //   font-weight: 700;
  //   margin-bottom: 8px;
  //   color: #222222;
  // `,
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
    const [isLoading, setIsLoading] = useState<boolean>(false);
    // const [server, setServer] = useState<string>('');
    // const [port, setPort] = useState<string>('');
    const [connected, setConnected] = useState<boolean>(false);

    // 使用上下文
    const appContext = useAppContext();
    // const { computerInfo ,server,port} = useAppContext();
    const { computerInfo } = useAppContext();

    const reloadTip = () => {
      setIsLoading(false);
      setConnected(false);
      Modal.confirm({
        title: 'Information',
        icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
        content: 'Connection failed. Please try again.',
        okText: 'Confirm',
        cancelText: 'Close',
        centered: true,
        okButtonProps: {
          style: { backgroundColor: '#0052cc',width: '90px' ,display: 'none'}
        },
        cancelButtonProps: {
          style: { width: '90px' }
        }
      })
    }

    const oaServerTip = () => {
      setIsLoading(false);
      setConnected(false);
      Modal.confirm({
        title: 'Information',
        icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
        content: 'Failed to obtain OA Server information; installation is not possible.',
        okText: 'Confirm',
        cancelText: 'Close',
        centered: true,
        okButtonProps: {
          style: { backgroundColor: '#0052cc',width: '90px' ,display: 'none'}
        },
        cancelButtonProps: {
          style: { width: '90px' }
        }
      })
    }

     const seedLabelTip = (content:String) => {
      setIsLoading(false);
      setConnected(false);
      Modal.confirm({
        title: 'Information',
        icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
        content: content,
        okText: 'Confirm',
        cancelText: 'Close',
        centered: true,
        okButtonProps: {
          style: { backgroundColor: '#0052cc',width: '90px' ,display: 'none'}
        },
        cancelButtonProps: {
          style: { width: '90px' }
        },
        onCancel: () => {
          (window as any).runtime?.Quit();
        }
      })
    }

  // 添加连接服务器的函数
  const connectToServer = async () => {
    // 检查 server 和 port 是否已输入
    // if (!server || !port) {
    //   Modal.confirm({
    //     title: 'Information',
    //     icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
    //     content: 'Please input the server and port.',
    //     cancelText: 'Close',
    //     centered: true,
        
    //     okButtonProps: {
    //       style: { backgroundColor: '#0052cc',width: '90px' ,display: 'none'}
    //     },
    //     cancelButtonProps: {
    //       style: { width: '90px' }
    //     }
    //   });
    //   return;
    // }

    setIsLoading(true);

    try {
      // 初始化客户端连接
      // const info = await InitClient();
      // appContext.setServer(server);
      // appContext.setPort(port);
      
      getOAServer();
      
    } catch (error) {
      reloadTip();
    } 
    
  };

  // 获取OAServer
  const getOAServer = async () => {
    try {
      const oaServer = await GetOAServer(appContext.computerInfo.ip);
      if (typeof oaServer === 'string' && oaServer.length > 0) {
        // appContext.setComputerInfo({
        //   name: computerInfo.name,
        //   seed: "-",
        //   oa:  oaServer,
        //   ip: computerInfo.ip,
        // });
        getSeedlabel(oaServer);
      }else{
        oaServerTip();
      }
      
    } catch (error) {
      oaServerTip();
    } 
  };

    // 获取Seedlabel
  const getSeedlabel = async (oaServer:string) => {
    try {
      const seed = await GetSeedLabel();
      if (typeof seed.seedlabel === 'string' && seed.seedlabel.length > 0) {
        if (seed.status === 'Active') {        
          checkSeedlabel(seed.seedlabel);
        }else{
          seedLabelTip('The seedlabel on this computer has expired. Please reinstall the system before running this program.');
        }
      }else{
        seedLabelTip('Failed to obtain seedlabel; installation is not possible.');
      }
      appContext.setComputerInfo({
        name: computerInfo.name,
        seed: seed.seedlabel,
        oa:  oaServer,
        ip: computerInfo.ip,
      });
    } catch (error) {
      seedLabelTip('The seedlabel on this computer is invalid and the program cannot be run.');
    } 
  };

  const checkSeedlabel = async (seed:string) => {
    try {
      const result = await CheckSeedLabel();
      if (result) {
        getPrinterModels();
      }else{
        seedLabelTip('The seedlabel on this computer is invalid.');
      }      
    } catch (error) {
      seedLabelTip('The seedlabel on this computer is invalid.');
    }   
  };

  // 获取打印机品牌
  const getPrinterModels = async () => {
    try {
      const models = await GetPrinterModels();
      if (models !== null) {
          // 保存到上下文中
         appContext.setPrinterModels(models as any[]);
      
         getNetworkPinterList();
      }else{
        reloadTip();
      }
    } catch (error) {
      reloadTip();
    } 
  };

  // 获取网络打印机
  const getNetworkPinterList = async () => {
    try {
      const models = await GetNetworkPinterList("");
      if (models !== null) {
        // 保存到上下文中
        appContext.setNetworkPinterModels(models as any[]);
        setIsLoading(false);
        setConnected(true);
      }else{
        reloadTip();
      }
    } catch (error) {
      reloadTip();
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
      
      {/* <div className={styles.inputContainer}>
        <div className={styles.inputGroup}>
          <span className={styles.inputLabel}>Server:</span>
          <Input 
            placeholder="Please input" 
            value={server}
            onChange={(e) =>  appContext.setServer(e.target.value)}
            style={{ width: '230px' }}
            disabled={connected}
          />
        </div>
        <div className={styles.inputGroup}>
          <span className={styles.inputLabel}>Port:</span>
          <Input 
            placeholder="Please input" 
            value={port}
            onChange={(e) =>  appContext.setPort(e.target.value)}
            style={{ width: '230px' }}
            disabled={connected}
          />
        </div>
      </div> */}
      <Text className={styles.connectingText}>
        {isLoading ? "Connecting to the server, please wait..." : (connected ? `Connected successfully` : " ")}
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