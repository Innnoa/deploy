import {useEffect, useState} from 'react';
import './App.css';
import 'antd/dist/reset.css';
import { Modal } from 'antd';
import { ExclamationCircleFilled } from '@ant-design/icons';
import Info from './components/info';
import Welcome from './components/welcome';
import Configuration from './components/configuration';
import Deploy from './components/deploy';
import { AppProvider } from './context/AppContext';
import { GetStartPage } from '../wailsjs/go/main/App'; // 导入Go绑定方法
import { EventsOn, EventsOff } from "../wailsjs/runtime/runtime";
import { CancelInatallation } from "../wailsjs/go/deploy/Deploy";

const App: React.FC = () => {
    //0: welcome, 1: configuration, 2: deploy
    // 根据URL参数初始化状态
    const [showComponentsPage, setShowComponentsPage] = useState(0);
    const [startPage, setStartPage] = useState("")

    useEffect(() => {
        // 全局监听窗口关闭事件
        const handleGlobalClose = () => {
            Modal.confirm({
                title: 'Confirm Exit',
                icon: <ExclamationCircleFilled style={{ color: '#faad14' }} />,
                content: 'Are you sure you want to exit the application?',
                okText: 'Exit',
                cancelText: 'Cancel',
                centered: true,
                okButtonProps: {
                    style: { backgroundColor: '#0052cc', width: '90px' }
                },
                cancelButtonProps: {
                    style: { width: '90px' }
                },
                onOk: () => {
                    CancelInatallation();
                }
            });
        };

        EventsOn("onBeforeClose", handleGlobalClose);

        return () => {
            EventsOff("onBeforeClose");
        };
    }, []);

    const getStartPage = async () => {
        const page = await GetStartPage();
        setStartPage(page)
        if (page === 'deploy') {
            setShowComponentsPage(2)
        }
    };  

    const showWelcomePage = () => {
        setShowComponentsPage(0);
    };
    const showConfigurationPage = () => {
        setShowComponentsPage(1);
    };
    const showDeployPage = () => {
        setShowComponentsPage(2);
    };

    getStartPage()

    return (
        <AppProvider>
            <div id="App" style={{ display: 'flex', width: '100%', height: '100vh' }}>
                <Info/>
                {showComponentsPage === 0 && (
                    <Welcome onStartClick={showConfigurationPage} />
                )}
                <div style={{ display: showComponentsPage === 1 ? 'block' : 'none' , width: '78%'}}>
                    <Configuration onBack={showWelcomePage} onSwitchToDeploy={showDeployPage}/>
                </div>
                {showComponentsPage === 2 && (
                    <Deploy onDeployBack={showConfigurationPage} startPage={startPage}/>
                )}
            </div>
        </AppProvider>
        
    )
}

export default App