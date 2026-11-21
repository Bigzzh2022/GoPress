import { useRef, useState } from 'react';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { App, Button, Card, Empty, Input, Pagination, Segmented, Skeleton } from 'antd';
import {
  CloudUploadOutlined,
  CopyOutlined,
  CustomerServiceOutlined,
  DeleteOutlined,
  FileOutlined,
  InboxOutlined,
  LinkOutlined,
  PictureOutlined,
  SearchOutlined,
  VideoCameraOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import type { MediaItem, MediaKind, MediaQuery } from '../../types';

const kindOptions = [
  { value: '', label: '全部' },
  { value: 'image', label: '图片' },
  { value: 'video', label: '视频' },
  { value: 'audio', label: '音频' },
  { value: 'other', label: '其他' },
];

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function kindOf(item: MediaItem): MediaKind {
  if (item.mime_type.startsWith('image/')) return 'image';
  if (item.mime_type.startsWith('video/')) return 'video';
  if (item.mime_type.startsWith('audio/')) return 'audio';
  return 'other';
}

function extensionOf(name: string) {
  const index = name.lastIndexOf('.');
  return index > -1 ? name.slice(index + 1).toUpperCase() : 'FILE';
}

export default function MediaPage() {
  useDocumentTitle('媒体库');
  const { message, modal } = App.useApp();
  const queryClient = useQueryClient();
  const inputRef = useRef<HTMLInputElement>(null);
  const [filters, setFilters] = useState<MediaQuery>({ page: 1, page_size: 24, keyword: '', kind: '' });
  const [keyword, setKeyword] = useState('');
  const [dragging, setDragging] = useState(false);
  const [uploading, setUploading] = useState(false);

  const mediaQuery = useQuery({ queryKey: ['admin', 'media', filters], queryFn: () => adminApi.media(filters), placeholderData: keepPreviousData });
  const deleteMutation = useMutation({
    mutationFn: adminApi.deleteMedia,
    onSuccess: async () => {
      message.success('文件已删除');
      await queryClient.invalidateQueries({ queryKey: ['admin', 'media'] });
    },
    onError: (error) => message.error(getErrorMessage(error)),
  });

  const uploadFiles = async (files: File[]) => {
    if (!files.length || uploading) return;
    setUploading(true);
    let failed = 0;
    for (const file of files) {
      try {
        await adminApi.uploadMedia(file);
      } catch (error) {
        failed += 1;
        message.error(`${file.name}：${getErrorMessage(error)}`);
      }
    }
    setUploading(false);
    setFilters((current) => ({ ...current, page: 1 }));
    await queryClient.invalidateQueries({ queryKey: ['admin', 'media'] });
    if (failed < files.length) message.success(`已上传 ${files.length - failed} 个文件`);
  };

  const copyLink = async (item: MediaItem) => {
    try {
      await navigator.clipboard.writeText(new URL(item.url, window.location.origin).href);
      message.success('文件链接已复制');
    } catch {
      message.error('复制失败，请手动打开文件后复制地址');
    }
  };
  const confirmDelete = (item: MediaItem) => modal.confirm({
    title: '删除这个文件？',
    content: `「${item.original_name}」将从服务器移除，已引用该地址的文章会出现失效链接。`,
    okText: '删除', cancelText: '取消', okButtonProps: { danger: true },
    onOk: () => deleteMutation.mutateAsync(item.id).then(() => undefined),
  });

  const changeFilter = (changes: Partial<MediaQuery>) => setFilters((current) => ({ ...current, page: 1, ...changes }));

  return (
    <div className="admin-page media-page">
      <div className="admin-page-heading"><div><p>ASSET LIBRARY</p><h1>媒体库</h1><span>上传并管理文章使用的图片、视频、附件等任意格式文件。</span></div></div>
      <div
        className={`media-dropzone${dragging ? ' is-dragging' : ''}`}
        onDragOver={(event) => { event.preventDefault(); setDragging(true); }}
        onDragLeave={() => setDragging(false)}
        onDrop={(event) => { event.preventDefault(); setDragging(false); void uploadFiles(Array.from(event.dataTransfer.files)); }}
        onClick={() => inputRef.current?.click()}
        role="button"
        tabIndex={0}
        onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') inputRef.current?.click(); }}
      >
        <InboxOutlined />
        <strong>{uploading ? '正在上传…' : '拖拽文件到这里，或点击选择'}</strong>
        <span>支持任意格式，可一次选择多个文件；上传后即可复制链接</span>
        {uploading && <Skeleton.Button active size="small" />}
        {!uploading && <Button icon={<CloudUploadOutlined />} onClick={(event) => { event.stopPropagation(); inputRef.current?.click(); }}>选择文件</Button>}
      </div>
      <input ref={inputRef} type="file" multiple hidden onChange={(event) => { const files = Array.from(event.target.files || []); event.target.value = ''; void uploadFiles(files); }} />

      <Card className="admin-table-card media-card">
        <div className="media-toolbar">
          <Segmented value={filters.kind || ''} onChange={(value) => changeFilter({ kind: value as MediaQuery['kind'] })} options={kindOptions} />
          <div className="media-search">
            <Input value={keyword} onChange={(event) => setKeyword(event.target.value)} onPressEnter={() => changeFilter({ keyword })} prefix={<SearchOutlined />} placeholder="搜索文件名或类型" allowClear />
            <Button type="primary" ghost icon={<SearchOutlined />} onClick={() => changeFilter({ keyword })}>查询</Button>
          </div>
        </div>

        {mediaQuery.isLoading ? <div className="media-skeleton"><Skeleton active paragraph={{ rows: 5 }} /></div> : !mediaQuery.data?.items.length ? (
          <Empty className="media-empty" image={Empty.PRESENTED_IMAGE_SIMPLE} description={filters.keyword || filters.kind ? '没有匹配的文件' : '媒体库还是空的'} />
        ) : (
          <div className="media-grid">
            {mediaQuery.data.items.map((item) => {
              const kind = kindOf(item);
              return (
                <article className="media-item" key={item.id}>
                  <a className="media-preview" href={item.url} target="_blank" rel="noreferrer" aria-label={`打开 ${item.original_name}`}>
                    {kind === 'image' ? <img src={item.url} alt={item.original_name} loading="lazy" /> : (
                      <div className={`media-file-icon ${kind}`}>
                        {kind === 'video' ? <VideoCameraOutlined /> : kind === 'audio' ? <CustomerServiceOutlined /> : <FileOutlined />}
                        <span>{extensionOf(item.original_name)}</span>
                      </div>
                    )}
                  </a>
                  <div className="media-info">
                    <strong title={item.original_name}>{item.original_name}</strong>
                    <span>{formatSize(item.size)} · {dayjs(item.created_at).format('YYYY-MM-DD')}</span>
                  </div>
                  <div className="media-actions">
                    <Button type="text" size="small" icon={<CopyOutlined />} onClick={() => void copyLink(item)} aria-label={`复制 ${item.original_name} 链接`} />
                    <a className="ant-btn ant-btn-text ant-btn-sm" href={item.url} target="_blank" rel="noreferrer" aria-label={`新窗口打开 ${item.original_name}`}><LinkOutlined /></a>
                    <Button type="text" size="small" danger icon={<DeleteOutlined />} onClick={() => confirmDelete(item)} aria-label={`删除 ${item.original_name}`} />
                  </div>
                </article>
              );
            })}
          </div>
        )}

        {!!mediaQuery.data?.total && (
          <div className="media-pagination">
            <span>共 {mediaQuery.data.total} 个文件</span>
            <Pagination current={mediaQuery.data.page} pageSize={filters.page_size} total={mediaQuery.data.total} showSizeChanger onChange={(page, page_size) => setFilters((current) => ({ ...current, page, page_size }))} />
          </div>
        )}
      </Card>
    </div>
  );
}
