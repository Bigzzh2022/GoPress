import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Alert, Button, Card, List, Skeleton, Tag } from 'antd';
import { CheckCircleOutlined, CloudSyncOutlined, GithubOutlined, ReloadOutlined, WarningOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';

function formatTime(value: string) {
  if (!value || value === 'unknown') return '未知';
  const parsed = dayjs(value);
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm') : value;
}

export default function SystemUpdatePage() {
  useDocumentTitle('检查更新');
  const [refreshToken, setRefreshToken] = useState(0);
  const updateQuery = useQuery({
    queryKey: ['admin', 'system-update', refreshToken],
    queryFn: () => adminApi.checkUpdate(refreshToken > 0),
    staleTime: 5 * 60 * 1000,
    refetchInterval: 5 * 60 * 1000,
    retry: false,
  });
  const data = updateQuery.data;
  const recheck = () => setRefreshToken((current) => current + 1);

  const state = !data
    ? null
    : !data.comparable
      ? { className: 'is-unknown', title: '无法比较当前构建', description: '当前构建没有记录可匹配的提交信息，下面仍会显示分支上的最新提交。' }
      : data.update_available
        ? { className: 'is-behind', title: `发现 ${data.behind} 个新提交`, description: '服务器上的版本落后于远程分支，可以按下方步骤拉取新镜像。' }
        : { className: 'is-current', title: '已是最新版本', description: '当前部署与远程分支的代码一致。' };

  return (
    <div className="admin-page">
      <div className="admin-page-heading">
        <div><p>MAINTENANCE</p><h1>检查更新</h1><span>对比当前部署与 {data?.branch || 'main'} 分支的最新代码，只读检查，不会自动改动服务器。</span></div>
        <Button icon={<ReloadOutlined />} loading={updateQuery.isFetching} onClick={recheck}>重新检查</Button>
      </div>

      {updateQuery.isError ? (
        <Alert
          type="error" showIcon message="检查更新失败" description={getErrorMessage(updateQuery.error)}
          action={<Button size="small" onClick={recheck}>重试</Button>}
        />
      ) : updateQuery.isLoading ? (
        <Card className="admin-card update-card"><Skeleton active paragraph={{ rows: 5 }} /></Card>
      ) : data && state ? (
        <>
          <div className="update-grid">
            <Card className="admin-card update-card" title="当前部署">
              <dl className="update-facts">
                <div><dt>版本</dt><dd>{data.current.version || 'dev'}</dd></div>
                <div>
                  <dt>提交</dt>
                  <dd>
                    {data.current.short_commit || 'unknown'}
                    {data.current.commit && data.current.commit !== 'unknown' && (
                      <a href={`https://github.com/${data.repository}/commit/${data.current.commit}`} target="_blank" rel="noreferrer">源码</a>
                    )}
                  </dd>
                </div>
                <div><dt>构建时间</dt><dd>{formatTime(data.current.build_time)}</dd></div>
              </dl>
            </Card>
            <Card className="admin-card update-card" title="检查目标">
              <dl className="update-facts">
                <div><dt>仓库</dt><dd><a href={`https://github.com/${data.repository}`} target="_blank" rel="noreferrer"><GithubOutlined /> {data.repository}</a></dd></div>
                <div><dt>分支</dt><dd>{data.branch}</dd></div>
                <div><dt>检查时间</dt><dd>{formatTime(data.checked_at)}{data.cached && <Tag className="update-cached">缓存</Tag>}</dd></div>
              </dl>
            </Card>
          </div>

          <Card className="admin-card update-card">
            <div className={`update-banner ${state.className}`}>
              <span className="update-banner-icon">{state.className === 'is-current' ? <CheckCircleOutlined /> : <CloudSyncOutlined />}</span>
              <div>
                <strong>{state.title}</strong>
                <p>{state.description}</p>
              </div>
              {data.comparable && data.update_available && data.compare_url && (
                <Button href={data.compare_url} target="_blank" rel="noreferrer">查看差异</Button>
              )}
            </div>

            {data.latest && (
              <div className="update-latest">
                <div>
                  <span className="update-label">分支最新提交</span>
                  <strong>{data.latest.message || '(无提交说明)'}</strong>
                  <div className="update-meta">
                    <Tag>{data.latest.short_sha}</Tag>
                    <span>{data.latest.author || '未知作者'}</span>
                    <span>{formatTime(data.latest.date)}</span>
                  </div>
                </div>
                {data.latest.url && <a href={data.latest.url} target="_blank" rel="noreferrer">在 GitHub 查看</a>}
              </div>
            )}
          </Card>

          {!!data.commits.length && (
            <Card className="admin-card update-card" title={`待更新提交（${data.commits.length}）`}>
              <List
                size="small" dataSource={data.commits}
                renderItem={(commit) => (
                  <List.Item extra={commit.url ? <a href={commit.url} target="_blank" rel="noreferrer">查看</a> : null}>
                    <div className="update-commit-row">
                      <Tag>{commit.short_sha}</Tag>
                      <span className="update-commit-message">{commit.message}</span>
                      <small>{commit.author} · {formatTime(commit.date)}</small>
                    </div>
                  </List.Item>
                )}
              />
            </Card>
          )}

          <Card className="admin-card update-card" title="如何应用更新">
            <div className="update-steps">
              <p><WarningOutlined /> 推送到 {data.branch} 后，GitHub Actions 会自动运行测试并推送新镜像。检查功能不会自动重启服务。</p>
              <ol>
                <li>等待 GitHub Actions 的镜像构建完成。</li>
                <li>在部署服务器上拉取并重启：</li>
              </ol>
              <pre className="update-command">{`docker compose pull\ndocker compose up -d`}</pre>
              <p className="update-hint">镜像默认使用 latest 标签；如需固定版本，可在 .env 中设置 GOPRESS_TAG 为某个提交标签。</p>
            </div>
          </Card>
        </>
      ) : null}
    </div>
  );
}
