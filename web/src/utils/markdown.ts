import type { ReactNode } from 'react';

export interface TocItem {
  id: string;
  text: string;
  level: 2 | 3;
}

export function nodeText(node: ReactNode): string {
  if (typeof node === 'string' || typeof node === 'number') return String(node);
  if (Array.isArray(node)) return node.map(nodeText).join('');
  if (node && typeof node === 'object' && 'props' in node) {
    const props = node.props as { children?: ReactNode };
    return nodeText(props.children);
  }
  return '';
}

export function headingId(text: string) {
  return text
    .trim()
    .toLowerCase()
    .replace(/[`*_~[\]()]/g, '')
    .replace(/[^\p{Letter}\p{Number}\s-]/gu, '')
    .replace(/\s+/g, '-') || 'section';
}

export function extractHeadings(markdown: string): TocItem[] {
  const headings: TocItem[] = [];
  const withoutCode = markdown.replace(/```[\s\S]*?```/g, '');
  for (const match of withoutCode.matchAll(/^(#{2,3})\s+(.+)$/gm)) {
    const text = match[2].replace(/\[([^\]]+)\]\([^)]*\)/g, '$1').replace(/[`*_~]/g, '').trim();
    headings.push({ id: headingId(text), text, level: match[1].length as 2 | 3 });
  }
  return headings;
}

export function readingTime(markdown: string) {
  const plain = markdown.replace(/[#>*_`~[\]()!-]/g, '').trim();
  const chinese = plain.match(/[\u3400-\u9fff]/g)?.length || 0;
  const words = plain.replace(/[\u3400-\u9fff]/g, ' ').split(/\s+/).filter(Boolean).length;
  return Math.max(1, Math.ceil(chinese / 400 + words / 220));
}
