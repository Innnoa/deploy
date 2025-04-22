import React from 'react';
import { Space } from 'antd';
import { createStyles } from 'antd-style';
import logo from '../../assets/images/info-logo.png';
import computerImg from '../../assets/images/computer-img.png';
import seedName from '../../assets/images/seed-name.png';
import serverName from '../../assets/images/server-name.png';
import ipName from '../../assets/images/ip-name.png';

// 使用 createStyles 定义样式
const useStyles = createStyles(({ css }) => ({
    logo: css`
    display: flex;
    justify-content: left;
    margin-top: calc(64 * 100vh / 1024);
    margin-bottom: 18px;
    width: calc(98 * 100vh / 1024);
    height: calc(98 * 100vh / 1024);
  `,
  infoContainer: css`
    background-color: #013FBF;
    color: white;
    padding: 20px;
    height: 100%;
    min-height: 100vh;
     width: 27%; /* 设置宽度为总宽度的27% */
  `,

  titleW: css`
  color: white;
  text-align: left;
  display: block;
  font-size: calc(34 * 100vh / 1024);
  `,
  titleD: css`
  color: white;
  text-align: left;
  display: block;
  font-size: calc(42 * 100vh / 1024);
  font-weight: 900;
  `,
  titleI: css`
  color: white;
  text-align: left;
  display: block;
  font-size: calc(24 * 100vh / 1024);
  margin-top: 40px;
  margin-bottom: 10px;
  `,

  infoItem: css`
    display: flex;
    align-items: center;
    margin-bottom: 1px;
    background-color: rgba(244, 245, 247, 0.05);
    padding: 10px;
    border-radius: 4px;
    height: calc(100 * 100vh / 1024);
  `,
  infoIcon: css`
    background-color: rgba(255, 255, 255, 0.2);
    width: calc(48 * 100vh / 1024);
    height: calc(48 * 100vh / 1024);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 15px;
    margin-left: 15px;
    font-size: 20px;
  `,
  infoContent: css`
    display: flex;
    flex-direction: column;
  `,
  infoText: css`
    color: white;
    margin-bottom: 5px;
    text-align: left;
    font-size: calc(20 * 100vh / 1024);
    font-weight: 900;
  `,
  infoTextContent: css`
    color: white;
    text-align: left;
    font-size: calc(20 * 100vh / 1024);
  `
}));

interface InfoProps {
  computerName?: string;
  seed?: string;
  oaServer?: string;
  ipAddress?: string;
}

const Info: React.FC<InfoProps> = ({ 
  computerName = 'C81363', 
  seed = 'CW11V24B', 
  oaServer = 'HPFS3OABAH2', 
  ipAddress = '10.50.241.78' 
}) => {
  const { styles } = useStyles();

  return (
    <div className={styles.infoContainer}>
        <img src={logo} className={styles.logo} alt="logo" />

        <span className={styles.titleW} >Welcome,</span>
        <span className={styles.titleD} >Deploy</span>
        <span className={styles.titleI} >Information</span>
      
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <div className={styles.infoItem}>
          <img src={computerImg} className={styles.infoIcon} alt="logo" />
          <div className={styles.infoContent}>
            <div className={styles.infoText}>Computer Name</div>
            <div className={styles.infoTextContent}> {computerName}</div>
          </div>
        </div>
        
        <div className={styles.infoItem}>
          <img src={seedName} className={styles.infoIcon} alt="logo" />
          <div className={styles.infoContent}>
            <div className={styles.infoText}>Seed</div>
            <div className={styles.infoTextContent}>{seed}</div>
          </div>
        </div>
        
        <div className={styles.infoItem}>
         <img src={serverName} className={styles.infoIcon} alt="logo" />
          <div className={styles.infoContent}>
            <div className={styles.infoText}>OA Server</div>
            <div className={styles.infoTextContent}>{oaServer}</div>
          </div>
        </div>
        
        <div className={styles.infoItem}>
        <img src={ipName} className={styles.infoIcon} alt="logo" />
          <div className={styles.infoContent}>
            <div className={styles.infoText}>IP Address</div>
            <div className={styles.infoTextContent}>{ipAddress}</div>
          </div>
        </div>
      </Space>
    </div>
  );
};

export default Info;