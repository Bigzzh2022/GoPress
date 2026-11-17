import { useEffect, useRef } from 'react';
import Vditor from 'vditor';
import 'vditor/dist/index.css';

interface VditorEditorProps {
  value?: string;
  onChange?: (value: string) => void;
}

export function VditorEditor({ value = '', onChange }: VditorEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const instanceRef = useRef<Vditor | null>(null);
  const initialValueRef = useRef(value);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  useEffect(() => {
    if (!containerRef.current) return;
    let disposed = false;
    const editor = new Vditor(containerRef.current, {
      height: 620,
      minHeight: 480,
      mode: 'ir',
      placeholder: '从一个想法开始，用 Markdown 记录你的思考…',
      cache: { enable: false },
      // Serve Vditor's runtime assets from our own origin instead of the default
      // unpkg CDN; third-party scripts would otherwise run inside the admin
      // session, where they could call the publishing API.
      cdn: `${import.meta.env.BASE_URL}vditor`,
      counter: { enable: true, type: 'text' },
      toolbarConfig: { pin: true },
      toolbar: [
        'headings', 'bold', 'italic', 'strike', '|',
        'quote', 'list', 'ordered-list', 'check', 'code', 'inline-code', 'link', 'table', 'line', '|',
        'undo', 'redo', '|', 'edit-mode', 'both', 'preview', 'outline', 'fullscreen',
      ],
      preview: { delay: 250, mode: 'both' },
      after: () => {
        if (disposed) return;
        if (initialValueRef.current) editor.setValue(initialValueRef.current);
      },
      input: (content: string) => onChangeRef.current?.(content),
    });
    instanceRef.current = editor;

    return () => {
      disposed = true;
      try {
        instanceRef.current?.destroy();
      } catch {
        // Vditor may still be initializing when a route changes.
      }
      instanceRef.current = null;
    };
  }, []);

  return <div className="vditor-shell" ref={containerRef} aria-label="Markdown 文章编辑器" />;
}
