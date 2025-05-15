import {useState} from 'react';
import './App.css';
import 'antd/dist/reset.css';
import Info from './components/info';
import Welcome from './components/welcome';
import Configuration from './components/configuration';
import Deploy from './components/deploy';
import { AppProvider } from './context/AppContext';

const App: React.FC = () => {
    //0: welcome, 1: configuration, 2: deploy
    const [showComponentsPage, setShowComponentsPage] = useState(0);

    const showWelcomePage = () => {
        setShowComponentsPage(0);
    };
    const showConfigurationPage = () => {
        setShowComponentsPage(1);
    };
    const showDeployPage = () => {
        setShowComponentsPage(2);
    };

    return (
        <AppProvider>
            <div id="App" style={{ display: 'flex', width: '100%', height: '100vh' }}>
                <Info />
                {showComponentsPage === 0 && (
                    <Welcome onStartClick={showConfigurationPage} />
                )}
                {showComponentsPage === 1 && (
                    <Configuration onBack={showWelcomePage} onSwitchToDeploy={showDeployPage}/>
                )}
                {showComponentsPage === 2 && (
                    <Deploy onDeployBack={showConfigurationPage}/>
                )}
            </div>
        </AppProvider>
        
    )
}

export default App