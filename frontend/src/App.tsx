import {useState} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Greet} from "../wailsjs/go/main/App";
import { PlusOutlined } from '@ant-design/icons';
import 'antd/dist/reset.css';
import { Button, Space, Typography, Card } from 'antd';
import Info from './components/info';
import Welcome from './components/welcome';

const App: React.FC = () => {
    const [resultText, setResultText] = useState("Please enter your name below 👇");
    const [name, setName] = useState('');
    const updateName = (e: any) => setName(e.target.value);
    const updateResultText = (result: string) => setResultText(result);

    // 显示信息组件
    const [showInfo, setShowInfo] = useState(false);

    function greet() {
        Greet(name).then(updateResultText);
    }

    return (
        <div id="App" style={{ display: 'flex', width: '100%', height: '100vh' }}>
            <Info />
            <Welcome />
        </div>
        // <div id="App">
        //     {showInfo ? (
        //         <Info />
        //     ) : (
        //         <>
        //             <img src={logo} id="logo" alt="logo"/>
        //             <div id="result" className="result">{resultText}</div>
        //             <div id="input" className="input-box">
        //                 <input id="name" className="input" onChange={updateName} autoComplete="off" name="input" type="text"/>
        //                 <Button type="primary" onClick={greet}>Greet</Button>
        //             </div>
        //             <Button 
        //                 type="primary" 
        //                 style={{ marginTop: '20px' }} 
        //                 onClick={() => setShowInfo(true)}
        //             >
        //                 显示信息
        //             </Button>
        //         </>
        //     )}
        // </div>
    )
}

export default App