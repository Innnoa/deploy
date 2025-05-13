import React, { useEffect, useState } from 'react';
import { Button, Typography } from 'antd';
import { createStyles } from 'antd-style';
import computerLogo from '../../assets/images/computer-logo.png';

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
  startButton: css`
    background-color: #0052cc;
    width: 150px;
    height: 40px;
  `
}));
interface WelcomeProps {
    onStartClick: () => void;
}
const Welcome: React.FC<WelcomeProps> = ({ onStartClick }) => {
    const { styles } = useStyles();
    const [packages, setPackages] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  useEffect(() => {
    // 调用后端的 GetAllPackages 方法
    const fetchPackages = async () => {
      try {
        // 使用 window.go.main.App 访问后端导出的方法
        const allPackages = await window.go.main.App.GetAllPackages();
        setPackages(allPackages);
        console.log('获取所有包信息成功:', allPackages);
      } catch (error) {
        console.error('获取包信息失败:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchPackages();
  }, []);

  return (
    <div className={styles.welcomeContainer}>
      <div className={styles.monitorImage}>
        <img src={computerLogo} alt="Computer" className={styles.computerLogo} />
      </div>
      
      <span className={styles.title} >Deploy Tool for</span>
      <span className={styles.subtitle} >Windows</span>
      
      {/* <Text className={styles.connectingText}>Connecting to the Servers, please wait...</Text> */}
      <Text className={styles.connectingText}>
      {isLoading ? "正在连接服务器，请稍候..." : `已连接，找到 ${packages.length} 个包`}
    </Text>
      <Button type="primary" className={styles.startButton} onClick={onStartClick}>
        Start
      </Button>
    </div>
  );
};

export default Welcome;