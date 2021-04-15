import React, { useEffect } from 'react';

import {
  Result,
  Typography
} from 'antd';

import {
    CloseCircleTwoTone
} from '@ant-design/icons';

const { Title } = Typography;

const Error404 = () => {

    useEffect(() => {
      document.title = "404 Not Found | Factom Anchor Platform";
    }, []);

    return (
      <Result
          icon={<CloseCircleTwoTone />}
          title={<Title level={2}>404 Not Found</Title>}
      />
    );
};

export default Error404;
