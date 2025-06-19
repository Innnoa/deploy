import {useEffect, useState} from 'react';
import './App.css';
import 'antd/dist/reset.css';
import Info from './components/info';
import Welcome from './components/welcome';
import Configuration from './components/configuration';
import Deploy from './components/deploy';
import { AppProvider } from './context/AppContext';
import { GetStartPage } from '../wailsjs/go/main/App'; // 导入Go绑定方法

const App: React.FC = () => {
    //0: welcome, 1: configuration, 2: deploy
    // 根据URL参数初始化状态
    const [showComponentsPage, setShowComponentsPage] = useState(0);

    const getStartPage = async () => {
        const page = await GetStartPage();
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
                <Info />
                {showComponentsPage === 0 && (
                    <Welcome onStartClick={showConfigurationPage} />
                )}
                <div style={{ display: showComponentsPage === 1 ? 'block' : 'none' , width: '78%'}}>
                    <Configuration onBack={showWelcomePage} onSwitchToDeploy={showDeployPage}/>
                </div>
                {showComponentsPage === 2 && (
                    <Deploy onDeployBack={showConfigurationPage}/>
                )}
            </div>
        </AppProvider>
        
    )
}

export default App