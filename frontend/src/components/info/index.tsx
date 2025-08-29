import React, { useEffect, useState }  from 'react';
import { Space ,Modal} from 'antd';
import { createStyles } from 'antd-style';
import logo from '../../assets/images/info-logo.png';
import computerImg from '../../assets/images/computer-img.png';
import seedName from '../../assets/images/seed-name.png';
import serverName from '../../assets/images/server-name.png';
import ipName from '../../assets/images/ip-name.png';
import {GetComputerInfo , IsAdmin, CheckNewVersion} from "../../../wailsjs/go/deploy/Deploy";
import { useAppContext } from '../../context/AppContext';
import { ExclamationCircleFilled} from '@ant-design/icons';


//合并冲突
//git pull --no-rebase origin main

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
  margin-top: calc(40 * 100vh / 1024);
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
    margin-left: 10px;
    font-size: 20px;
  `,
  infoContent: css`
    display: flex;
    flex-direction: column;
  `,
  infoText: css`
    color: white;
    text-align: left;
    font-size: calc(20 * 100vh / 1024);
    font-weight: 900;
    line-height: calc(24 * 100vh / 1024);
  `,
  infoTextContent: css`
    color: white;
    text-align: left;
    font-size: calc(20 * 100vh / 1024);
    line-height: calc(24 * 100vh / 1024);
  `
}));

const Info: React.FC = () => {
  const { styles } = useStyles();
  const { computerInfo } = useAppContext();
  // 使用上下文
  const appContext = useAppContext();


  useEffect(() => {
    // 调用后端的 GetComputerInfo 方法
    const fetchComputerInfo = async () => {
      try {
        
        const info = await GetComputerInfo();
         appContext.setComputerInfo({
          name: info.name,
          seed: info.seed === '' ? '-' : info.seed,
          oa: info.oa === '' ? '-' : info.oa,
          ip: info.ip
        });
      } catch (error) {
        // console.error('获取计算机信息失败:', error);
      }
    };

    fetchComputerInfo();
  }, []);

  useEffect(() => {
    //管理员才可以继续
    const fetchIsAdmin = async () => {
      try {
        
        const info = await IsAdmin();
          if (info != true){
            Modal.confirm({
              title: 'Information',
              icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
              content: 'Please start the program as an administrator user.',
              cancelText: 'Close',
              centered: true,
              
              okButtonProps: {
                style: { display: 'none'}
              },
              cancelButtonProps: {
                style: { width: '90px' }
              },
              onCancel: () => {
                (window as any).runtime?.Quit();
              }
            });
          } else {
            checkNewVersion();
          }

      } catch (error) {
      }
    };

    const checkNewVersion = async () => {
      const info = await CheckNewVersion();
      if (info == true){
        Modal.confirm({
          title: 'Information',
          icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
          content: 'There is a new version available, please use the new version.',
          cancelText: 'Close',
          centered: true,
          
          okButtonProps: {
            style: { display: 'none'}
          },
          cancelButtonProps: {
            style: { width: '90px' }
          },
          onCancel: () => {
            (window as any).runtime?.Quit();
          }
        });
      } 
    };

    fetchIsAdmin();


    
  }, []);  

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
            <div className={styles.infoTextContent}> {computerInfo.name}</div>
          </div>
        </div>
        
        <div className={styles.infoItem}>
          <img src={seedName} className={styles.infoIcon} alt="logo" />
          <div className={styles.infoContent}>
            <div className={styles.infoText}>Seed</div>
            <div className={styles.infoTextContent}>{computerInfo.seed}</div>
          </div>
        </div>
        
        <div className={styles.infoItem}>
         <img src={serverName} className={styles.infoIcon} alt="logo" />
          <div className={styles.infoContent}>
            <div className={styles.infoText}>OA Server</div>
            <div className={styles.infoTextContent}>{computerInfo.oa}</div>
          </div>
        </div>
        
        <div className={styles.infoItem}>
        <img src={ipName} className={styles.infoIcon} alt="logo" />
          <div className={styles.infoContent}>
            <div className={styles.infoText}>IP Address</div>
            <div className={styles.infoTextContent}>{computerInfo.ip}</div>
          </div>
        </div>
      </Space>
    </div>
  );
};

export default Info;