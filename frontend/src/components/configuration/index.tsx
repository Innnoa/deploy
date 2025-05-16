import React, { useState } from 'react';
import { Select, Input, Button, Checkbox, Divider, Table, Modal } from 'antd';
import { SearchOutlined ,ExclamationCircleFilled} from '@ant-design/icons';
import { createStyles } from 'antd-style';
import type { ColumnsType } from 'antd/es/table';
import { useAppContext } from '../../context/AppContext';
import { GetPrinterDrivers } from "../../../wailsjs/go/deploy/Deploy"; 

const useStyles = createStyles(({ css }) => ({
  configContainer: css`
    width: 78%;
    padding: 22px;
    background-color: #F4F5F7;
    height: 100vh;
    position: relative;
  `,
  section: css`
    background: white;
    padding: 20px 20px 15px 20px;
    border-radius: 4px;
    margin-bottom: 16px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  `,
  printerSection: css`
  background: white;
  padding: 20px;
  border-radius: 4px;
//   margin-bottom: 4px; 
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  // height: calc(100vh - 330px); /* 动态计算最大高度 */
  // overflow-y: auto;
`,
  sectionTitle: css`
    font-size: 20px;
    font-weight: 700;
    margin-bottom: 20px;
    color: #222222;
    text-align: left;
  `,
  formItem: css`
    margin-bottom: 16px;
    display: flex;
    align-items: center;
  `,
  label: css`
    width: 120px;
    margin-right: 16px;
    color: #222222;
    text-align: left;
  `,
  divider: css`
    margin: 16px 0;
    border-top: 1px dashed #E8E8E8;
  `,
  searchInput: css`
    margin-bottom: 16px;
    background-color: #f5f5f7;
    border-radius: 7px;
    height: 35px;
  `,
  customTable: css`
    // background-color: #f5f5f7;
    
    .ant-table-thead > tr > th {
      background-color: #f5f5f7 !important;
    }
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
    color: #777;
    // margin-left: 22px;
    font-size: 15px;
    margin-left: auto;
    margin-right: 16px;
    text-align: right;
  `,
  footerLink: css`
    color: #013FBF;
    // cursor: pointer;
    font-size: 15px;
  `,
  buttonGroup: css`
    display: flex;
    gap: 12px;
  `,
  startButton: css`
    background-color: #0052cc;
    &:hover {
    background-color: #013FBF !important;
  }
  `
}));

interface PrinterData {
  key: string;
  polNo: string;
  ip: string;
}
interface ConfigurationProps {
    onBack: () => void; // 添加返回按钮回调函数
    onSwitchToDeploy: () => void; // 新增
  }

const Configuration: React.FC<ConfigurationProps> = ({ onBack ,onSwitchToDeploy}) => {
  const { styles } = useStyles();
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [printerModel, setPrinterModel] = useState<string | undefined>();
  const [printerDriver, setPrinterDriver] = useState<string | undefined>();
  const [driverOptions, setDriverOptions] = useState<{value: string, label: string}[]>([]);

  // 使用上下文获取数据
  const { printerModels, server, port } = useAppContext();

  const printerData: PrinterData[] = [
    { key: '1', polNo: 'P123141', ip: '10.50.234.180' },
    { key: '2', polNo: 'P123141', ip: '10.50.234.180' },
    { key: '3', polNo: 'P123141', ip: '10.50.234.180' },
    { key: '4', polNo: 'P123141', ip: '10.50.234.180' },
    { key: '5', polNo: 'P123141', ip: '10.50.234.180' },
    { key: '6', polNo: 'P123141', ip: '10.50.234.180' },
    { key: '7', polNo: 'P123141', ip: '10.50.234.180' },
    // { key: '8', polNo: 'P123141', ip: '10.50.234.180' },
    // { key: '9', polNo: 'P123141', ip: '10.50.234.180' },
    // { key: '10', polNo: 'P123141', ip: '10.50.234.180' },


    

  ];

   // 将 printerModels 转换为 Select 组件需要的 options 格式
   const options = (printerModels || []).map(model => ({
    value: model,
    label: model
  }));

  const columns: ColumnsType<PrinterData> = [
    {
      title: 'Pol No',
      dataIndex: 'polNo',
      key: 'polNo',
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      key: 'ip',
    },
  ];

  const onSelectChange = (newSelectedRowKeys: React.Key[]) => {
    setSelectedRowKeys(newSelectedRowKeys);
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectChange,
  };
  // 动态计算是否需要滚动
  const getTableScroll = () => {
    const rowHeight = 47; // 每行高度（根据实际情况调整）
    const headerHeight = 47; // 表头高度（根据实际情况调整）
    const contentHeight = printerData.length * rowHeight + headerHeight ;
    console.log('contentHeight =',contentHeight);
    
    const maxHeight = window.innerHeight - 470;
    console.log('maxHeight =',maxHeight);
    
    // 只有当内容高度超过最大高度时才启用滚动
    return contentHeight > maxHeight ? { y: 'calc(100vh - 510px)' } : {};
  };
  const handleNext = () => {
    console.log('打印机型号:', printerModel);
    console.log('打印机驱动:', printerDriver);
    console.log('选中的打印机:', selectedRowKeys);
    if (!printerModel && !printerDriver && selectedRowKeys.length === 0) {
      Modal.confirm({
        title: 'Information',
        icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
        content: 'No printer selected, are you sure to start deploy?',
        okText: 'Confirm',
        cancelText: 'Cancel',
        centered: true,
        okButtonProps: {
          style: { backgroundColor: '#0052cc' }
        },
        onOk: () => {
          // 用户确认后的操作
          onSwitchToDeploy(); // 切换到Deploy组件
          console.log('用户确认继续');
          // 这里可以添加继续的逻辑
        }
      });
    } else {
      // 正常继续的逻辑
      onSwitchToDeploy(); // 切换到Deploy组件
      console.log('继续下一步');
    }
  };
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
    // 选择打印机型号时调用驱动接口
    const handlePrinterModelChange = async (value: string) => {
      setPrinterModel(value);
      try {
        const drivers = await GetPrinterDrivers(value);
         // 将drivers数组中的每个对象的app_name提取出来作为选项
          const driverOpts = Array.isArray(drivers) 
          ? drivers.map((driver: any) => ({ 
              value: driver.app_name, 
              label: driver.app_name 
            }))
          : [];
        setDriverOptions(driverOpts);
        // 如果有驱动选项，自动选择第一个
        if (driverOpts.length > 0) {
          setPrinterDriver(driverOpts[0].value);
        }else{
          setPrinterDriver(undefined);
        }
      } catch (e) {
        setDriverOptions([]);
        setPrinterDriver(undefined);
      }
    };

  return (
    <div className={styles.configContainer}>
      {/* 本地打印机配置 */}
      <div className={styles.section}>
        <div className={styles.sectionTitle}>Local Printer Configuration</div>
        <div className={styles.formItem}>
          <span className={styles.label}>Printer Models:</span>
          <Select style={{ width: 350, textAlign: 'left' }} placeholder="select" 
            options={options} dropdownStyle={{ textAlign: 'left' }}
            onChange={handlePrinterModelChange}
            />
        </div>
        <Divider dashed className={styles.divider} />
        <div className={styles.formItem}>
          <span className={styles.label}>Printer Driver:</span>
          <Select style={{ width: 350, textAlign: 'left'  }} placeholder="select" 
            options={driverOptions} dropdownStyle={{ textAlign: 'left' }}
            onChange={(value) => setPrinterDriver(value)}
            value={printerDriver}
            />
        </div>
      </div>

      {/* Print Q Configuration */}
      <div className={styles.printerSection}>
        <div className={styles.sectionTitle}>Print Q Configuration</div>
        <Input
          prefix={<SearchOutlined />}
          placeholder="Search for your Printer"
          className={styles.searchInput}
          variant="borderless"
          allowClear
        />

        <Table 
          className={styles.customTable}
          rowSelection={rowSelection}
          columns={columns}
          dataSource={printerData}
          pagination={false}
          size="middle"
          bordered={false}
            // scroll={{ y: 'calc(100vh - 510px)' }} /* 设置表格内部滚动 */
          scroll={getTableScroll() || undefined}
        />
      </div>

      {/* 底部按钮 */}
      <div className={styles.footer}>
        <div className={styles.footerText}>
          Additional Printer Driver Installation.
          <span className={styles.footerLink}>Click NEXT to Next step</span>
        </div>
        <div className={styles.buttonGroup}>
          <Button onClick={onBack}>Back</Button>
          <Button danger onClick={handleCancel}>Cancel</Button>
          <Button type="primary" onClick={handleNext}
                  className={styles.startButton}>NEXT</Button>
        </div>
      </div>
    </div>
  );
};

export default Configuration;