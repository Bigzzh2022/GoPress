import { ArrowLeft, Compass } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useDocumentTitle } from '../../app/hooks';

export default function NotFoundPage() {
  useDocumentTitle('页面未找到');
  return (
    <main className="not-found page-container">
      <div className="not-found-number">4<span><Compass /></span>4</div>
      <p className="eyebrow">LOST IN THE HYPERTEXT</p>
      <h1>这一页不在路由里</h1>
      <p>链接可能已经改变，或者你想找的内容从未存在。<br />别担心，好的文章还在原来的地方。</p>
      <Link to="/" className="button button-dark"><ArrowLeft size={17} /> 回到首页</Link>
    </main>
  );
}
