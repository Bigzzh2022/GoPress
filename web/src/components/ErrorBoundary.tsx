import { Component, type ErrorInfo, type ReactNode } from 'react';
import { CircleAlert } from 'lucide-react';

interface Props { children: ReactNode }
interface State { hasError: boolean }

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    if (import.meta.env.DEV) console.error('页面渲染异常', error, info);
  }

  render() {
    if (this.state.hasError) {
      return (
        <main className="fatal-error" role="alert">
          <CircleAlert size={38} aria-hidden="true" />
          <p className="eyebrow">SOMETHING WENT WRONG</p>
          <h1>这一页暂时没有准备好</h1>
          <p>页面遇到了意外问题。你可以刷新重试，或回到首页继续阅读。</p>
          <div className="flex flex-wrap justify-center gap-3">
            <button className="button button-dark" onClick={() => window.location.reload()}>刷新页面</button>
            <a className="button button-outline" href="/">返回首页</a>
          </div>
        </main>
      );
    }
    return this.props.children;
  }
}
