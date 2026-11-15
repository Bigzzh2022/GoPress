import { useEffect, useState } from 'react';
import { ArrowUpRight, Menu, Search, X } from 'lucide-react';
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom';

const navItems = [
  { to: '/', label: '首页', end: true },
  { to: '/posts', label: '文章' },
  { to: '/about', label: '关于' },
];

export function PublicLayout() {
  const [menuOpen, setMenuOpen] = useState(false);
  const location = useLocation();

  useEffect(() => {
    setMenuOpen(false);
    window.scrollTo({ top: 0, behavior: 'instant' });
  }, [location.pathname]);

  return (
    <div className="public-shell">
      <a className="skip-link" href="#main-content">跳到正文</a>
      <header className="site-header">
        <div className="site-header-inner">
          <Link to="/" className="brand" aria-label="GoPress 首页">
            <span className="brand-mark">G<span /></span>
            <span className="brand-word">GoPress<small>CODE & CURIOSITY</small></span>
          </Link>
          <nav className="desktop-nav" aria-label="主导航">
            {navItems.map((item) => (
              <NavLink key={item.to} to={item.to} end={item.end} className={({ isActive }) => isActive ? 'active' : ''}>
                {item.label}
              </NavLink>
            ))}
          </nav>
          <div className="header-actions">
            <Link to="/posts" className="icon-button" aria-label="搜索文章"><Search size={19} /></Link>
            <Link to="/posts" className="header-cta">开始阅读 <ArrowUpRight size={16} /></Link>
            <button className="mobile-menu-button" onClick={() => setMenuOpen((value) => !value)} aria-expanded={menuOpen} aria-controls="mobile-navigation" aria-label={menuOpen ? '关闭导航' : '打开导航'}>
              {menuOpen ? <X size={22} /> : <Menu size={22} />}
            </button>
          </div>
        </div>
        {menuOpen && (
          <nav id="mobile-navigation" className="mobile-nav" aria-label="移动端导航">
            {navItems.map((item, index) => (
              <NavLink key={item.to} to={item.to} end={item.end}>
                <span>0{index + 1}</span>{item.label}<ArrowUpRight size={18} />
              </NavLink>
            ))}
          </nav>
        )}
      </header>
      <main id="main-content"><Outlet /></main>
      <footer className="site-footer">
        <div className="footer-main page-container">
          <div>
            <Link to="/" className="brand footer-brand"><span className="brand-mark">G<span /></span><span className="brand-word">GoPress</span></Link>
            <p>写清晰的代码，保持诚实的好奇。<br />关于 Go、云原生与软件工程的独立记录。</p>
          </div>
          <div className="footer-links">
            <div><strong>阅读</strong><Link to="/posts">全部文章</Link><Link to="/posts?category=backend">后端工程</Link><Link to="/about">关于作者</Link></div>
            <div><strong>连接</strong><a href="mailto:hello@gopress.dev">邮件联系</a><Link to="/admin/login">内容管理</Link><a href="#main-content">回到顶部</a></div>
          </div>
        </div>
        <div className="footer-bottom page-container"><span>© {new Date().getFullYear()} GoPress</span><span>BUILT WITH GO · DESIGNED FOR READING</span></div>
      </footer>
    </div>
  );
}
