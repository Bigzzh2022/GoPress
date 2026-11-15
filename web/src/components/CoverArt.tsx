import { Braces, Database, Layers3, Radio, TerminalSquare } from 'lucide-react';

const icons = [TerminalSquare, Braces, Database, Layers3, Radio];

export function CoverArt({ seed, label = 'GOPRESS', large = false }: { seed: number; label?: string; large?: boolean }) {
  const variant = Math.abs(seed) % 5;
  const Icon = icons[variant];
  return (
    <div className={`cover-art cover-${variant} ${large ? 'cover-large' : ''}`} aria-hidden="true">
      <span className="cover-index">NO. {String(Math.abs(seed) + 1).padStart(3, '0')}</span>
      <span className="cover-orbit" />
      <Icon className="cover-icon" strokeWidth={1.15} />
      <span className="cover-label">{label}</span>
      <span className="cover-lines" />
    </div>
  );
}
