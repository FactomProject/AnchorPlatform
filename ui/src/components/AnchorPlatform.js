import React, { useState, useEffect, useRef } from 'react';
import { Router, Route, Link, Switch } from 'react-router-dom';

import { Menu, Layout, Input, Badge, Divider, Select, Typography, message, Form, Tooltip, Dropdown, Button } from 'antd';

// import { isValidAddress } from 'factom';

import { IconContext } from "react-icons";
import {
    RiCheckboxMultipleBlankLine, RiLinksLine, RiVipDiamondLine, RiServerLine, RiCodeSSlashLine, RiAnchorLine, RiExternalLinkLine, RiDashboardLine, RiTerminalBoxLine, RiSettings3Line, RiBracesLine
} from 'react-icons/ri';

import {
    LoadingOutlined, DownOutlined
} from '@ant-design/icons';

import axios from 'axios';
import Moment from 'react-moment';

import Logo from './common/Logo';
import ScrollToTop from './common/ScrollToTop';
import History from './common/History';

import { NotifyNetworkError } from './common/Notifications';

import Dashboard from './sections/Dashboard';
import Error404 from './sections/Error404';


const { Header, Content, Footer } = Layout;
const { Search } = Input;
const { Text } = Typography;

const AnchorPlatform = props => {
    
  const [currentMenu, setCurrentMenu] = useState([window.location.pathname]);

  const handleMenuClick = e => {
    if (e.key === "logo") {
        setCurrentMenu("/dashboard");
    } else {
        setCurrentMenu([e.key]);
    }
  };

  const handleChangelogClick = () => {
    setCurrentMenu(null);
  };

  useEffect(() => {
    if (window.location.pathname === "/" || window.location.pathname.includes("dashboard")) {
        setCurrentMenu("/dashboard");
    }
    if (window.location.pathname.includes("anchormakers")) {
        setCurrentMenu("/anchormakers");
    }
    if (window.location.pathname.includes("receipts")) {
        setCurrentMenu("/receipts");
    }
    if (window.location.pathname.includes("receipts")) {
        setCurrentMenu("/settings");
    }
  }, []);

  return (
    <Router history={History}>
    <ScrollToTop />
        <Layout>
        <Header style={{ padding: 0, margin: 0 }}>
            <Menu theme="dark" mode="horizontal" onClick={handleMenuClick} selectedKeys={currentMenu}>
                <Menu.Item key="logo" className="menu-no-hover">
                    <Link to="/">
                        <Logo />
                    </Link>
                </Menu.Item>
                <Menu.Item key="/dashboard">
                    <Link to="/">
                        <IconContext.Provider value={{ className: 'react-icons' }}><RiDashboardLine /></IconContext.Provider>
                        <span className="nav-text">Dashboard</span>
                    </Link>
                </Menu.Item>
                <Menu.Item key="/anchormakers">
                    <Link to="/anchormakers">
                    <IconContext.Provider value={{ className: 'react-icons' }}><RiAnchorLine /></IconContext.Provider>
                        <span className="nav-text">Anchor Makers</span>
                    </Link>
                </Menu.Item>
                <Menu.Item key="/receipts">
                    <Link to="/receipts">
                    <IconContext.Provider value={{ className: 'react-icons' }}><RiBracesLine /></IconContext.Provider>
                        <span className="nav-text">Receipts</span>
                    </Link>
                </Menu.Item>
                <Menu.Item key="/settings">
                    <Link to="/settings">
                    <IconContext.Provider value={{ className: 'react-icons' }}><RiSettings3Line /></IconContext.Provider>
                        <span className="nav-text">Settings</span>
                    </Link>
                </Menu.Item>
            </Menu>
        </Header>
        <Content style={{ padding: '20px', margin: 0 }}>
            <Switch>
                <Route exact path="/" component={Dashboard} />
                <Route exact path="/dashboard" component={Dashboard} />
                <Route exact path="/anchormakers" component={Dashboard} />
                <Route exact path="/receipts" component={Dashboard} />
                <Route exact path="/settings" component={Dashboard} />
                <Route component={Error404} />
            </Switch>
        </Content>
        <Footer>
            Version: 0.1.0
        </Footer>
        </Layout>
    </Router>
  );
};

export default AnchorPlatform;
