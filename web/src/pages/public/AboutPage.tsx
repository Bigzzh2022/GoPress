import { useQuery } from '@tanstack/react-query';
import { ArrowRight, Braces, Coffee, Compass, Mail, MapPin } from 'lucide-react';
import { Link } from 'react-router-dom';
import { blogApi } from '../../api/blog';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import { ErrorState, LoadingSkeleton } from '../../components/Feedback';

export default function AboutPage() {
  useDocumentTitle('关于');
  const profileQuery = useQuery({ queryKey: ['public', 'profile'], queryFn: blogApi.profile });
  const settingsQuery = useQuery({ queryKey: ['public', 'settings'], queryFn: blogApi.settings, staleTime: 5 * 60 * 1000 });
  const profile = profileQuery.data;
  const siteName = settingsQuery.data?.site_name || 'GoPress';
  const contactEmail = settingsQuery.data?.contact_email || 'hello@gopress.dev';

  return (
    <div className="about-page">
      <section className="about-hero page-container">
        <div className="about-intro">
          <p className="eyebrow">ABOUT THE PERSON BEHIND THE SCREEN</p>
          <h1>工程师，<br />写作者，<br /><span>长期主义者。</span></h1>
          <p className="about-lead">我相信好的软件和好的文字有相同的标准：<br />清晰、克制，并且经得起时间。</p>
        </div>
        <div className="profile-card">
          {profileQuery.isLoading ? <LoadingSkeleton cards={1} /> : profileQuery.isError ? <ErrorState message={getErrorMessage(profileQuery.error)} onRetry={() => profileQuery.refetch()} /> : (
            <>
              <div className="profile-monogram" aria-label={`${profile?.name || siteName} 的文字头像`}>{profile?.name?.slice(0, 1) || siteName.slice(0, 1)}<span /></div>
              <p className="eyebrow">AUTHOR & MAINTAINER</p>
              <h2>{profile?.name || `${siteName} 作者`}</h2>
              <p>{profile?.bio || '专注于 Go 与云原生技术，记录真实世界里的软件工程。'}</p>
              <div className="profile-details"><span><MapPin size={15} />中国 · 远程工作</span><span><Coffee size={15} />咖啡与终端爱好者</span></div>
              <a className="button button-dark" href={`mailto:${contactEmail}`}><Mail size={16} /> 和我聊聊</a>
            </>
          )}
        </div>
      </section>

      <section className="about-story page-container">
        <div className="story-label"><span>01</span><p>WHY I WRITE</p></div>
        <div className="story-content"><h2>写作，是另一种<br />理解世界的方式。</h2><div><p>{siteName} 始于一个简单的想法：把那些散落在终端、issue 和笔记本里的经验，整理成可以被检索、被讨论，也可以被时间检验的内容。</p><p>这里不追求每日更新，也不复制文档。我更愿意分享一次真实故障的排查、一个架构取舍的过程，或是一个让代码更清晰的微小决定。</p></div></div>
      </section>

      <section className="principles-section">
        <div className="page-container"><div className="section-heading light"><div><p className="eyebrow">WORKING PRINCIPLES</p><h2>我在意的事情</h2></div></div><div className="principle-grid">
          <div><Braces /><span>01</span><h3>简单优于巧妙</h3><p>可维护的系统来自清晰的边界、明确的取舍，以及足够无聊的解决方案。</p></div>
          <div><Compass /><span>02</span><h3>理解先于工具</h3><p>工具会更替，问题却不会。先理解约束与上下文，再选择合适的工具。</p></div>
          <div><Coffee /><span>03</span><h3>持续而非冲刺</h3><p>阅读、实践、写作和复盘，让进步成为一套可以长期运行的系统。</p></div>
        </div></div>
      </section>

      <section className="about-cta page-container"><p className="eyebrow">START SOMEWHERE</p><h2>从一个问题开始，<br />一起把复杂变简单。</h2><Link to="/posts" className="button button-ember">进入文章档案 <ArrowRight size={17} /></Link></section>
    </div>
  );
}
