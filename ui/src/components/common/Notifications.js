import React from 'react';
import { message } from 'antd';

export function NotifyNetworkError() {
  message.error('Anchor Platform API is unavailable');

  return <NotifyNetworkError />;
}