import React, { useState ,useEffect} from 'react';
import { Table, Button, Progress, Typography, Modal ,Tooltip} from 'antd';
import { ExclamationCircleFilled,CheckCircleFilled} from '@ant-design/icons';
import { createStyles } from 'antd-style';
import {GetInstallPackages,DoInstall,GetInstallStatus, Reboot, InstallAfterReboot, InitClient} from "../../../wailsjs/go/deploy/Deploy";

const { Text } = Typography;

const useStyles = createStyles(({ css }) => ({
  container: css`
    width: 78%;
    padding: 22px;
    background-color: #F4F5F7;
    height: 100vh;
    position: relative;
  `,
  progressBar: css`
    width: 100%;
    height: 62px;
    padding: 16px 24px 16px 24px;
    background: #fff;
    border-radius: 4px ;
    margin-bottom: 16px;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
     .ant-progress-outer {
      height: 20px !important;
      background: #f0f0f0;
      border-radius: 10px;
    }
      .ant-progress-bg{
      height: 20px !important;
      }
  `,
  tableCard: css`
    background: #fff;
    border-radius: 4px;
    box-shadow: 0 1px 3px rgba(0,0,0,0.03);
    padding: 20px;
  `,
  tableTitle: css`
    font-weight: 700;
    font-size: 16px;
    padding: 20px 24px 0 24px;
    color: #222;
  `,
  table: css`
    // margin: 24px;
    .ant-table-thead > tr > th {
      background-color: #f5f5f7 !important;
    }
  `,
  statusCompleted: css`
    color: #3CC91A;
    background: #f6ffed;
    border: 1px solid #3CC91A;
    border-radius: 4px;
    padding: 2px 12px;
    display: inline-block;
    width: 100px;
    text-align: center;
  `,
  statusFailed: css`
    color: #D8374D;
    background: #fff1f0;
    border: 1px solid #D8374D;
    border-radius: 4px;
    padding: 2px 12px;
    display: inline-block;
    width: 100px;
    text-align: center;
  `,
  statusRunning: css`
    color:#2055bf;
    background: #e6f4ff;
    border: 1px solid#CCCCCC;
    border-radius: 4px;
    padding: 2px 12px;
    display: inline-block;
    width: 100px;
    text-align: center;
  `,
  statusWaiting: css`
    color: #777777;
    background: #fafafa;
    border: 1px solid #CCCCCC;
    border-radius: 4px;
    padding: 2px 12px;
    display: inline-block;
    width: 100px;
    text-align: center;
  `,
  footer: css`
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 15px;
    position: absolute;
    bottom: 24px;
    left: 22px;
    right: 22px;
    background-color: #FFFFFF;
    border-radius: 4px;
    height: 60px; /* 添加固定高度以便计算间距 */
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  `,
  
  buttonGroup: css`
    margin-left: auto;
    display: flex;
    gap: 12px;
  `,
  startButton: css`
    width: 90px;
    background-color: #0052cc;
    &:hover {
    background-color: #013FBF !important;
  }
    &:disabled,
    &[disabled] {
      background-color: #CFCFCF !important;
      color: #fff !important;
      cursor: not-allowed !important;
    }
    &:disabled:hover,
    &[disabled]:hover {
      background-color: #CFCFCF !important;
      color: #fff !important;
      cursor: not-allowed !important;
    }
  `,
  backButton: css`
    width: 90px;
  }
  `,
}));

const columns = [
  {
    title: 'Task',
    dataIndex: 'appname',
    key: 'appname',
    ellipsis: true,
    render: (appname: string, record: any) => (
      <Tooltip title={appname} >
        <span>{appname}</span>
      </Tooltip>
    ),
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    width: 220,
    render: (status: string, record: any, index: number) => {
      if (status === 'Completed') {
        return <span className={record.styles.statusCompleted}>Completed</span>;
      }
      if (status === 'Failed') {
        return (
          <Tooltip title={record.error || 'Unknown error'}>
            <span className={record.styles.statusFailed}>Failed</span>
          </Tooltip>
        );
      }
      if (status === 'Running') {
        return <span className={record.styles.statusRunning}>Running</span>;
      }
      return <span className={record.styles.statusWaiting}>Waiting</span>;
    }
  }
];
    
  interface DeployProps {
    onDeployBack: () => void;
    startPage: string;
  }

  //退出程序
  const handleCancel = (finished:boolean) => {
    Modal.confirm({
      title: 'Confirm Exit',
      icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
      content: 'Are you sure you want to exit the application?',
      okText: 'Exit',
      cancelText: 'Cancel',
      centered: true,
      okButtonProps: {
        style: { backgroundColor: '#0052cc',width: '90px' }
      },
      cancelButtonProps: {
        style: { width: '90px' }
      },
      onOk: () => {
        if (finished == true) {
          console.log("exit app, reboot")
          Reboot();
        }

        (window as any).runtime?.Quit();
      }
    });
  };

const Deploy: React.FC<DeployProps> = ({ onDeployBack, startPage }) => {
  const { styles } = useStyles();
  const [allData, setAllData] = useState<any[]>([]);
  const [isRunning, setIsRunning] = useState(false);
  const intervalRef = React.useRef<number | null>(null);
  const [isDeployFinished, setIsDeployFinished] = useState(false);

  const getInstallPackages = async () => {
    try {
      const models = await GetInstallPackages();
      if (models !== null) {
        
        setAllData(models);
      }else{
        setAllData([]);
      }
    } catch (error) {
      setAllData([]);
    } 
  };

  //获取安装状态
  const getInstallStatus = async () => {
    try {
      const models = await GetInstallStatus();
      if (models !== null) {
        
        setAllData(models);
      }else{
        setAllData([]);
      }
    } catch (error) {
      setAllData([]);
    } 
  };

  //安装
  const doInstall = async () => {
    try {
      const status = await DoInstall();

      

       // 处理错误返回值 弹窗提示 -------------
      // if (typeof status === 'string' && status.length > 0) {
      //   Modal.error({
      //     title: 'Error',
      //     content: status,
      //     centered: true,
      //     okText: 'Confirm',
      //     okButtonProps: {
      //       style: { backgroundColor: '#0052cc' }
      //     },
      //   });
        
      //   if (intervalRef.current) {
      //     clearInterval(intervalRef.current);
      //     intervalRef.current = null;
      //   }
      //   setIsRunning(false);
      //   setIsDeployFinished(false);
      // }else{
      //   if (!intervalRef.current) {
      //     intervalRef.current = window.setInterval(() => {
      //       getInstallStatus();
      //     }, 1000); 
      //    }
      // }
      
    } catch (error) {
      
    } 
  };

  useEffect(() => {

    getInstallPackages();

    if (startPage === 'deploy') {
      console.log('InstallAfterReboot')
      setIsRunning(true);
      // 暂不处理错误返回值 --------------
      if (!intervalRef.current) {
        intervalRef.current = window.setInterval(() => {
          getInstallStatus();
        }, 1000); 
      }
      InstallAfterReboot();
    }
    
  }, []);

  // Run按钮点击事件
  const handleRun = () => {
    setIsRunning(true);

    doInstall();
    // 暂不处理错误返回值 --------------
    if (!intervalRef.current) {
      intervalRef.current = window.setInterval(() => {
        getInstallStatus();
      }, 1000); 
     }
  };
  // 组件卸载时清理定时器
  React.useEffect(() => {
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, []);
    

  // 处理数据
  const dataSource = allData.map(item => ({ ...item, styles }));

      // 动态计算是否需要滚动
  const getTableScroll = () => {
    const rowHeight = 60; // 每行高度（根据实际情况调整）
    const headerHeight = 47; // 表头高度（根据实际情况调整）
    const contentHeight = dataSource.length * rowHeight + headerHeight ;
    console.log('contentHeight =',contentHeight);
    
    const maxHeight = window.innerHeight - 255;
    console.log('maxHeight =',maxHeight);
    
    // 只有当内容高度超过最大高度时才启用滚动
    return contentHeight > maxHeight ? { y: 'calc(100vh - 295px)' } : {};
  };

   // 计算 percent
   const completedCount = dataSource.filter(item => item.status === 'Completed' || item.status === 'Failed').length;
   const percent = dataSource.length === 0 ? 0 : Math.round((completedCount / dataSource.length) * 100);
  // 监听 percent 达到 100 弹窗
  useEffect(() => {
    if (percent === 100 && isRunning) {
      setIsDeployFinished(true); // 部署完成
      Modal.confirm({
        title: 'Deploy complete!',
        icon: <CheckCircleFilled style={{ color: '#04B700' }} />,
        content: 'Exit the application?',
        okText: 'Exit',
        centered: true,
        okButtonProps: {
          style: { backgroundColor: '#0052cc',width: '90px' }
        },
        cancelButtonProps: {
          style: { width: '90px' }
        },
        onOk: () => {
          handleCancel(true);
        }
      });


      setIsRunning(false);
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }
  }, [percent, isRunning]);


return (
    <div className={styles.container}>
      {/* 顶部进度条 */}
      <div className={styles.progressBar}>
        <Progress 
            percent={percent}
            showInfo={false} 
            status={percent === 100 ? "success" : "active"}
            strokeColor={{
             '0%': '#CFCFCF',
             '50%': '#2055bf',
             '100%': '#013FBF'
             }}
            style={{ marginTop: 6 }}
          />
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginLeft:16, }}>
        <Text style={{ fontWeight: 400, whiteSpace: 'nowrap' }}>
          <span style={{ color: '#013FBF' }}>The {completedCount}</span>
          <span style={{ color: '#777777' }}> / Total of {dataSource.length}</span>
        </Text>
        </div>
      </div>

      {/* 列表卡片 */}
      <div className={styles.tableCard}>
        <Table
          className={styles.table}
          columns={columns}
          dataSource={dataSource}
          rowKey="id"
          pagination={false}
          scroll={getTableScroll() || undefined}
        />
      </div>

      {/* 底部按钮区 */}
      <div className={styles.footer}>

        <div className={styles.buttonGroup}>
          <Button onClick={onDeployBack}
            disabled={isRunning || isDeployFinished}
            className={styles.backButton}>Back</Button>
          <Button danger onClick={()=>handleCancel(isDeployFinished)}
            className={styles.backButton}>
            {isDeployFinished ? 'Close' : 'Cancel'}</Button>
          <Button type="primary" 
            className={styles.startButton}
            loading={isRunning}
            disabled={isRunning || dataSource.length === 0  || isDeployFinished}
            onClick={handleRun}>Run</Button>
        </div>
      </div>
    </div>
  );
};

export default Deploy;