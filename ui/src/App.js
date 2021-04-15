import React, { useState, useEffect } from 'react';
import { Spin } from 'antd';
import axios from 'axios';
import AnchorPlatform from './components/AnchorPlatform';
import './App.css';

axios.defaults.baseURL = process.env.REACT_APP_API_PATH;

const App = () => {
  const [loaded, setLoaded] = useState(false);

  const renderApp = () => {
    if (loaded) {
      return <AnchorPlatform />;
    } else {
      return <Spin size="large" className="loader" />;
    }
  };

  useEffect(() => {
    setLoaded(true);
  }, []);

  return <div>{renderApp()}</div>;
};

export default App;
