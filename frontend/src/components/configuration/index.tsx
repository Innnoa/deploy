import React from 'react';
import { Button, Typography } from 'antd';
import { createStyles } from 'antd-style';
import computerLogo from '../../assets/images/computer-logo.png';

const { Title, Text } = Typography;

const useStyles = createStyles(({ css }) => ({
  welcomeContainer: css`
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100vh;
    width: 70%;
    padding: 20px;
  `,
  monitorImage: css`
    width: 300px;
    height: 250px;
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
  `,
  subtitle: css`
    text-align: center;
    margin-bottom: 20px !important;
  `,
  connectingText: css`
    color: #8c8c8c;
    margin-bottom: 30px;
  `,
  startButton: css`
    background-color: #0052cc;
    width: 150px;
    height: 40px;
  `
}));

const Welcome: React.FC = () => {
  const { styles } = useStyles();

  return (
    <div className={styles.welcomeContainer}>
      <div className={styles.monitorImage}>
        <img src={computerLogo} alt="Computer" className={styles.computerLogo} />
      </div>
      
      <Title level={2} className={styles.title}>Deploy Tool for</Title>
      <Title level={2} className={styles.subtitle}>Windows</Title>
      
      <Text className={styles.connectingText}>Connecting to the Servers, please wait...</Text>
      
      <Button type="primary" className={styles.startButton}>
        Start
      </Button>
    </div>
  );
};

export default Welcome;