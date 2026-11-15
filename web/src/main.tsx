import '@ant-design/v5-patch-for-react-19';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClientProvider } from '@tanstack/react-query';
import { App as AntApp, ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import App from './App';
import { ErrorBoundary } from './components/ErrorBoundary';
import { queryClient } from './app/queryClient';
import 'highlight.js/styles/github.css';
import './styles.css';

dayjs.locale('zh-cn');

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <ConfigProvider
          locale={zhCN}
          theme={{
            token: {
              colorPrimary: '#d95d39',
              colorInfo: '#d95d39',
              colorText: '#262a27',
              colorBgLayout: '#f4f5f7',
              borderRadius: 8,
              fontFamily: 'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
            },
            components: {
              Layout: { headerBg: '#ffffff', siderBg: '#202521' },
              Menu: { darkItemBg: '#202521', darkItemSelectedBg: '#d95d39' },
              Button: { controlHeight: 38, fontWeight: 600 },
              Card: { headerFontSize: 16 },
            },
          }}
        >
          <AntApp message={{ maxCount: 3 }}>
            <App />
          </AntApp>
        </ConfigProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  </StrictMode>,
);
