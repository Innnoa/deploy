import React, { useState } from 'react';
import { Table, Button, Progress, Typography, Modal } from 'antd';
import { ExclamationCircleFilled} from '@ant-design/icons';
import { createStyles } from 'antd-style';

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
  `,
  tableTitle: css`
    font-weight: 700;
    font-size: 16px;
    padding: 20px 24px 0 24px;
    color: #222;
  `,
  table: css`
    margin: 0 24px;
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
    dataIndex: 'task',
    key: 'task',
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

  const allData = [
    { key: 1, task: 'Autologon_ITLU | Autologon For Training School', status: 'Completed' },
    { key: 2, task: 'Local Printer | Kyocera ECOSYS P3260dn...', status: 'Completed' },
    { key: 3, task: 'Printer Q | IP | 10.50.243.201', status: 'Failed' },
    { key: 4, task: 'Printer Q | Pol NO | P43688', status: 'Running' },
    { key: 5, task: 'Reborn | DeepFreeze Reborn Software', status: 'Waiting' },
    { key: 6, task: 'Reborn | DeepFreeze Reborn Software', status: 'Waiting' },
    { key: 7, task: 'Deploy | PC Deployment', status: 'Waiting' },
    { key: 8, task: 'Deploy | PC Deployment', status: 'Waiting' },
    // { key: 9, task: 'Deploy | PC Deployment', status: 'Waiting' },
    // { key: 10, task: 'Deploy | PC Deployment', status: 'Waiting' },
    // ...更多数据
  ];

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

  // 处理分页数据
  const dataSource = allData.slice((current - 1) * PAGE_SIZE, current * PAGE_SIZE)
    .map(item => ({ ...item, styles }));

return (
    <div className={styles.container}>
      {/* 顶部进度条 */}
      <div className={styles.progressBar}>
        <Progress percent={87} showInfo={false} status="active" 
            strokeColor={{
             '0%': '#CFCFCF',
             '50%': '#2055bf',
             '100%': '#013FBF'
             }}
            style={{ marginTop: 6 }}
          />
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginLeft:16, }}>
          <Text style={{ color: '#013FBF', fontWeight: 400 ,whiteSpace: 'nowrap'}}>The {current * PAGE_SIZE - PAGE_SIZE + 4} / Total of 15</Text>
        </div>
      </div>

      {/* 列表卡片 */}
      <div className={styles.tableCard}>
        <div className={styles.tableTitle}></div>
        <Table
          className={styles.table}
          columns={columns}
          dataSource={dataSource}
          rowKey="key"
          // scroll={{ y: 'calc(100vh - 340px)' }} /* 设置表格内部滚动 */
          // 根据数据源的总数来决定是否显示分页
          pagination={{
            current,
            pageSize: PAGE_SIZE,
            total: allData.length,
            showTotal: (total: number) => `Total ${total} items`,
            onChange: setCurrent,
            showSizeChanger: false,
            hideOnSinglePage: true,
          } }
        />
      </div>

      {/* 底部按钮区 */}
      <div className={styles.footer}>
        <span className={styles.footerText}>Click Cancel to stop deploy.</span>
        <div className={styles.buttonGroup}>
          <Button onClick={onDeployBack}>Back</Button>
          <Button danger onClick={handleCancel}>Cancel</Button>
          <Button type="primary" className={styles.startButton}>Run</Button>
        </div>
      </div>
    </div>
  );
};

export default Deploy;