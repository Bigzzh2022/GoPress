import ReactMarkdown, { defaultUrlTransform } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import type { Components } from 'react-markdown';
import { headingId, nodeText } from '../utils/markdown';

const components: Components = {
  h2: ({ children }) => <h2 id={headingId(nodeText(children))}>{children}</h2>,
  h3: ({ children }) => <h3 id={headingId(nodeText(children))}>{children}</h3>,
  a: ({ href = '', children }) => {
    const external = /^https?:\/\//.test(href);
    return (
      <a href={defaultUrlTransform(href)} target={external ? '_blank' : undefined} rel={external ? 'noreferrer noopener' : undefined}>
        {children}
      </a>
    );
  },
  input: ({ checked, type, ...props }) => type === 'checkbox'
    ? <input {...props} type="checkbox" checked={checked} readOnly disabled aria-label="任务列表项" />
    : <input {...props} type={type} readOnly />,
};

export function MarkdownContent({ content }: { content: string }) {
  return (
    <div className="article-body" data-testid="markdown-content">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[[rehypeHighlight, { detect: false, ignoreMissing: true }]]}
        components={components}
        skipHtml
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
