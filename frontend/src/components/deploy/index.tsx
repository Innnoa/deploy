import React, { useState ,useEffect} from 'react';
import { Table, Button, Progress, Typography, Modal } from 'antd';
import { ExclamationCircleFilled} from '@ant-design/icons';
import { createStyles } from 'antd-style';
import {GetInstallPackages,DoInstall,GetInstallStatus} from "../../../wailsjs/go/deploy/Deploy";

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
    border: 1px solid#2758ba;
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
  footerText: css`
    color: #013FBF;
    font-size: 15px;
    margin-left: auto;
    margin-right: 16px;
    text-align: right;
  `,
  buttonGroup: css`
    display: flex;
    gap: 16px;
  `,
  startButton: css`
    background-color: #0052cc;
    &:hover {
    background-color: #013FBF !important;
  }
  `
}));

const columns = [
  {
    title: 'Task',
    dataIndex: 'appname',
    key: 'appname',
    ellipsis: true,
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
        return <span className={record.styles.statusFailed}>Failed</span>;
      }
      if (status === 'Running') {
        return <span className={record.styles.statusRunning}>Running</span>;
      }
      return <span className={record.styles.statusWaiting}>Waiting</span>;
    }
  }
];

  // const allData = [
  //   { id: 1, appname: 'Autologon_ITLU | Autologon For Training School', status: 'Completed' },
  //   { id: 2, appname: 'Local Printer | Kyocera ECOSYS P3260dn...', status: 'Completed' },
  //   { id: 3, appname: 'Printer Q | IP | 10.50.243.201', status: 'Failed' },
  //   { id: 4, appname: 'Printer Q | Pol NO | P43688', status: 'Running' },
  //   { id: 5, appname: 'Reborn | DeepFreeze Reborn Software', status: 'Waiting' },
  //   { id: 6, appname: 'Reborn | DeepFreeze Reborn Software', status: 'Waiting' },
  //   { id: 7, appname: 'Deploy | PC Deployment', status: 'Waiting' },
  //   // { id: 8, appname: 'Deploy | PC Deployment', status: 'Waiting' },
  //   // { id: 9, appname: 'Deploy | PC Deployment', status: 'Waiting' },
  //   // { id: 10, appname: 'Deploy | PC Deployment111', status: 'Waiting' },
  //   // ...更多数据
  // ];

  const PAGE_SIZE = 7;

    
  interface DeployProps {
    onDeployBack: () => void;
  }

  //退出程序
  const handleCancel = () => {
    Modal.confirm({
      title: 'Confirm Exit',
      icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
      content: 'Are you sure you want to exit the application?',
      okText: 'Exit',
      cancelText: 'Cancel',
      centered: true,
      onOk: () => {
        (window as any).runtime?.Quit();
      }
    });
  };

const Deploy: React.FC<DeployProps> = ({ onDeployBack }) => {
  const { styles } = useStyles();
  const [current, setCurrent] = useState(1);
  const [allData, setAllData] = useState<any[]>([]);
  const [isRunning, setIsRunning] = useState(false);
  const intervalRef = React.useRef<NodeJS.Timeout | null>(null);

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

  const doInstall = async () => {
    try {
      const models = await DoInstall();
      
    } catch (error) {
      
    } 
  };

  useEffect(() => {
    
    getInstallPackages();
  }, []);

    // Run按钮点击事件
    const handleRun = () => {
      setIsRunning(true);
      
      doInstall();
      
      // 每隔一秒调用一次getInstallStatus
      if (!intervalRef.current) {
        intervalRef.current = setInterval(() => {
          getInstallStatus();
        }, 1000);
      }
    };

    // 组件卸载时清理定时器
    React.useEffect(() => {
      return () => {
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
        }
      };
    }, []);

  // 处理分页数据
  // const dataSource = allData.slice((current - 1) * PAGE_SIZE, current * PAGE_SIZE)
  //   .map(item => ({ ...item, styles }));
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
   const completedCount = dataSource.filter(item => item.status === 'Completed').length;
   const percent = dataSource.length === 0 ? 0 : Math.round((completedCount / dataSource.length) * 100);
  // 监听 percent 达到 100 弹窗
  useEffect(() => {
    if (percent === 100 && isRunning) {
      
      Modal.confirm({
        title: 'Information',
        icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
        content: 'Deploy complete!',
        okText: 'Confirm',
        centered: true,
        okButtonProps: {
          style: { backgroundColor: '#0052cc' }
        },
        onOk: () => {
          // 用户确认后的操作
          
          
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
            showInfo={false} status="active" 
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
        {/* <div className={styles.tableTitle}></div> */}
        <Table
          className={styles.table}
          columns={columns}
          dataSource={dataSource}
          rowKey="id"
          pagination={false}
          scroll={getTableScroll() || undefined}
          // scroll={{ y: 'calc(100vh - 340px)' }} /* 设置表格内部滚动 */
          // 根据数据源的总数来决定是否显示分页
          // pagination={{
          //   current,
          //   pageSize: PAGE_SIZE,
          //   total: allData.length,
          //   showTotal: (total: number) => `Total ${total} items`,
          //   onChange: setCurrent,
          //   showSizeChanger: false,
          //   hideOnSinglePage: true,
          // } }
        />
      </div>

      {/* 底部按钮区 */}
      <div className={styles.footer}>
        <span className={styles.footerText}>Click Cancel to stop deploy.</span>
        <div className={styles.buttonGroup}>
          <Button onClick={onDeployBack}
            disabled={isRunning}>Back</Button>
          <Button danger onClick={handleCancel}>Cancel</Button>
          <Button type="primary" 
            className={styles.startButton}
            loading={isRunning}
            onClick={handleRun}>Run</Button>
        </div>
      </div>
    </div>
  );
};

export default Deploy;