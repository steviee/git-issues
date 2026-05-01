import { useState } from 'react';
import { ChevronDown } from 'lucide-react';

interface Command {
  cmd: string;
  desc: string;
  example: string;
}

const commands: Command[] = [
  {
    cmd: 'issues init',
    desc: 'Initialize .issues/ directory in current repo',
    example: '$ issues init\nInitialized git-issues in .issues/',
  },
  {
    cmd: 'issues new',
    desc: 'Create with editor or inline flags',
    example: '$ issues new --title "Fix bug" --priority high --label bug\nCreated: .issues/0001-fix-bug.md (#1)',
  },
  {
    cmd: 'issues list',
    desc: 'Filter by status, priority, label; output as table/json/ids',
    example: '$ issues list --status open --priority critical\nID    PRI       STATUS       TITLE\n0001  critical  open         Fix auth bug',
  },
  {
    cmd: 'issues show',
    desc: 'Full detail with resolved relations and blocker indicators',
    example: '$ issues show 1\nIssue #1  ·  critical  ·  open\nFix auth bug',
  },
  {
    cmd: 'issues relate',
    desc: 'Bidirectional dependency sync',
    example: '$ issues relate 1 blocks 2\nRelated: #1 blocks #2\n         #2 depends-on #1 (auto)',
  },
  {
    cmd: 'issues board',
    desc: 'Interactive Kanban in the terminal',
    example: '$ issues board\n[Interactive TUI opens]',
  },
];

export default function CommandPreview() {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  const toggle = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <section id="commands" className="py-32 px-6 md:px-8 lg:px-12 border-t border-[#262626]">
      <div className="max-w-[1200px] mx-auto">
        <div className="text-center mb-16">
          <div className="inline-flex mb-4">
            <span className="font-mono text-xs font-medium uppercase tracking-[0.05em] text-[#525252] bg-[#141414] border border-[#262626] rounded-full px-3 py-1">
              Commands
            </span>
          </div>
          <h2 className="text-4xl md:text-[44px] font-bold leading-[1.1] tracking-[-0.02em]">
            Everything you need
          </h2>
        </div>

        <div className="max-w-[800px] mx-auto flex flex-col gap-3">
          {commands.map((command, index) => (
            <div
              key={command.cmd}
              className="rounded-xl border border-[#262626] overflow-hidden hover:border-[#525252] transition-colors"
            >
              <button
                onClick={() => toggle(index)}
                className="w-full flex items-center justify-between px-5 py-4 text-left"
              >
                <div className="flex items-center gap-3">
                  <span className="font-mono text-sm text-[#22c55e]">{command.cmd}</span>
                  <span className="text-sm text-[#a3a3a3]">{command.desc}</span>
                </div>
                <ChevronDown
                  size={18}
                  className={`text-[#525252] shrink-0 transition-transform ${openIndex === index ? 'rotate-180' : ''}`}
                />
              </button>
              {openIndex === index && (
                <div className="px-5 pb-4">
                  <div className="rounded-lg bg-[#0c0c0c] border border-[#1f1f1f] p-4 font-mono text-sm text-[#a3a3a3] whitespace-pre">
                    {command.example}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
