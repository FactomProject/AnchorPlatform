import React, { useState, useEffect } from 'react';
import axios from 'axios';

import {
  Typography,
  Statistic, Row, Col, Divider,
  message
} from 'antd';

import { NotifyNetworkError } from './../common/Notifications';


const { Title } = Typography;

const Dashboard = () => {

  const { heights, setHeights } = useState(null);
  const { fees, setFees } = useState(null);

  const getFees = async () => {
    try {
        const response = await axios.post('/v2', {
          jsonrpc: '2.0',
          id: 0,
          method: 'fees',
          params: {
          },
        }, {
          headers: {
            'Content-Type': 'application/json'
          }
        });
    }
    catch(error) {
        if (error.response) {
            message.error(error.response.data.error);
        } else {
            NotifyNetworkError();
        }
    }
  }

  const getHeights = async () => {
    try {
        const response = await axios.post('/v2', {
          jsonrpc: '2.0',
          id: 0,
          method: 'heights',
          params: {
          },
        }, {
          headers: {
            'Content-Type': 'application/json'
          }
        });
    }
    catch(error) {
        if (error.response) {
            message.error(error.response.data.error);
        } else {
            NotifyNetworkError();
        }
    }
  }

  useEffect(() => {
    document.title = "Dashboard | Factom Anchor Platform";
  }, []);

  useEffect(() => getFees(), []);
  useEffect(() => getHeights(), []);

  return (
    <div>
      <Title level={2}>Dashboard</Title>
      <Row gutter={24}>
        <Col span={4}>
            <Statistic title="DBHeight" value={heights ? heights.directoryblockheight : "…"} />
        </Col>
        <Col span={4}>
            <Statistic title="EBHeight" value={heights ? heights.entryblockheight : "…"} />
        </Col>
      </Row>
      <Divider></Divider>
      <Title level={3}>Ledgers</Title>
    </div>
  );

};

export default Dashboard;
