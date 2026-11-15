import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MarkdownContent } from '../components/MarkdownContent';

describe('MarkdownContent', () => {
  it('renders GFM and highlighted code without injecting raw HTML', () => {
    const { container } = render(
      <MarkdownContent content={'## 安全渲染\n\n| 语言 | 用途 |\n| --- | --- |\n| Go | 服务 |\n\n```go\nfmt.Println("hello")\n```\n\n<script>window.pwned = true</script>\n<img src=x onerror=alert(1)>'} />,
    );

    expect(screen.getByRole('heading', { name: '安全渲染' })).toHaveAttribute('id', '安全渲染');
    expect(screen.getByRole('table')).toBeInTheDocument();
    expect(screen.getByText('Go')).toBeInTheDocument();
    expect(container.querySelector('script')).not.toBeInTheDocument();
    expect(container.querySelector('img')).not.toBeInTheDocument();
    expect(container.querySelector('pre code')).toBeInTheDocument();
    expect((window as Window & { pwned?: boolean }).pwned).toBeUndefined();
  });

  it('sanitizes unsafe links and protects external links', () => {
    const { container } = render(<MarkdownContent content={'[危险链接](javascript:alert(1))\n\n[示例](https://example.com)'} />);
    const unsafe = screen.getByText('危险链接').closest('a');
    const external = screen.getByText('示例').closest('a');

    expect(unsafe?.getAttribute('href') || '').not.toMatch(/^javascript:/i);
    expect(external).toHaveAttribute('target', '_blank');
    expect(external).toHaveAttribute('rel', expect.stringContaining('noopener'));
    expect(container.querySelector('script')).not.toBeInTheDocument();
  });
});
