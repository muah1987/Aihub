import { useState, useEffect } from 'react';
import {
  knowledgeApi,
  type KnowledgeDocument,
  type DocumentChunk,
  type AutoExtraction,
  type SharedMemory,
  type MemoryVersion,
} from '../../api/knowledge';
import { memoryApi, type TeamMemory } from '../../api/memory';
import {
  FileText, Plus, Trash2, Search, X, Check, XCircle,
  ChevronRight, RotateCcw, Share2, Sparkles, BookOpen,
} from 'lucide-react';

interface KnowledgePanelProps {
  projectId: string;
}

type SubTab = 'documents' | 'versions' | 'shared' | 'extractions';

const extractionColors: Record<string, string> = {
  decision: 'bg-blue-500/20 text-blue-400',
  requirement: 'bg-green-500/20 text-green-400',
  architecture: 'bg-purple-500/20 text-purple-400',
  bug: 'bg-red-500/20 text-red-400',
  action_item: 'bg-yellow-500/20 text-yellow-400',
};

export function KnowledgePanel({ projectId }: KnowledgePanelProps) {
  const [subTab, setSubTab] = useState<SubTab>('documents');

  const tabs: { id: SubTab; label: string }[] = [
    { id: 'documents', label: 'Documents' },
    { id: 'versions', label: 'Versions' },
    { id: 'shared', label: 'Shared' },
    { id: 'extractions', label: 'Extractions' },
  ];

  return (
    <div className="flex flex-col h-full">
      <div className="flex gap-2 px-4 py-2 border-b border-[var(--color-border)] shrink-0 overflow-x-auto">
        {tabs.map(t => (
          <button key={t.id} onClick={() => setSubTab(t.id)}
            className={`px-3 py-1.5 text-xs rounded-lg whitespace-nowrap ${subTab === t.id ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg-secondary)] text-[var(--color-text-secondary)]'}`}>
            {t.label}
          </button>
        ))}
      </div>
      <div className="flex-1 overflow-hidden">
        {subTab === 'documents' && <DocumentsTab projectId={projectId} />}
        {subTab === 'versions' && <VersionsTab projectId={projectId} />}
        {subTab === 'shared' && <SharedTab projectId={projectId} />}
        {subTab === 'extractions' && <ExtractionsTab projectId={projectId} />}
      </div>
    </div>
  );
}

// ---- Documents Tab ----

function DocumentsTab({ projectId }: { projectId: string }) {
  const [docs, setDocs] = useState<KnowledgeDocument[]>([]);
  const [selectedDoc, setSelectedDoc] = useState<KnowledgeDocument | null>(null);
  const [chunks, setChunks] = useState<DocumentChunk[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<DocumentChunk[]>([]);
  const [loading, setLoading] = useState(true);

  // Create form
  const [title, setTitle] = useState('');
  const [fileName, setFileName] = useState('');
  const [fileType, setFileType] = useState('text');
  const [content, setContent] = useState('');

  useEffect(() => { loadDocs(); }, [projectId]);

  const loadDocs = async () => {
    setLoading(true);
    try {
      const res = await knowledgeApi.listDocuments(projectId);
      setDocs(res.data.documents || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  const handleCreate = async () => {
    if (!title.trim() || !content.trim()) return;
    try {
      await knowledgeApi.createDocument(projectId, {
        title, file_name: fileName || title, file_type: fileType, content,
      });
      setShowCreate(false);
      setTitle(''); setFileName(''); setContent('');
      loadDocs();
    } catch { /* empty */ }
  };

  const handleDelete = async (id: string) => {
    try {
      await knowledgeApi.deleteDocument(projectId, id);
      if (selectedDoc?.id === id) setSelectedDoc(null);
      loadDocs();
    } catch { /* empty */ }
  };

  const selectDoc = async (doc: KnowledgeDocument) => {
    setSelectedDoc(doc);
    try {
      const res = await knowledgeApi.getChunks(projectId, doc.id);
      setChunks(res.data.chunks || []);
    } catch { /* empty */ }
  };

  const handleSearch = async () => {
    if (!searchQuery.trim()) { setSearchResults([]); return; }
    try {
      const res = await knowledgeApi.searchChunks(projectId, searchQuery);
      setSearchResults(res.data.chunks || []);
    } catch { /* empty */ }
  };

  if (selectedDoc) {
    return (
      <div className="flex flex-col h-full">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-border)] shrink-0">
          <button onClick={() => setSelectedDoc(null)} className="text-sm text-[var(--color-primary)]">&larr; Back</button>
          <div className="flex-1">
            <h3 className="font-semibold text-sm">{selectedDoc.title}</h3>
            <span className="text-xs text-[var(--color-text-secondary)]">
              {selectedDoc.file_type} &middot; {selectedDoc.chunk_count} chunks &middot; {(selectedDoc.file_size / 1024).toFixed(1)} KB
            </span>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {chunks.map(chunk => (
            <div key={chunk.id} className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-mono text-[var(--color-text-secondary)]">Chunk #{chunk.chunk_index + 1}</span>
                <span className="text-xs text-[var(--color-text-secondary)]">{chunk.token_count} tokens</span>
              </div>
              <p className="text-sm whitespace-pre-wrap">{chunk.content}</p>
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div className="relative flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-secondary)]" />
          <input value={searchQuery} onChange={e => setSearchQuery(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && handleSearch()}
            placeholder="Semantic search across documents..."
            className="w-full pl-9 pr-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
        </div>
        <button onClick={() => setShowCreate(true)} className="flex items-center gap-1 px-3 py-2 bg-[var(--color-primary)] text-white rounded-lg text-xs hover:opacity-90">
          <Plus size={14} /> Add
        </button>
      </div>

      {searchResults.length > 0 && (
        <div className="border-b border-[var(--color-border)] p-4 space-y-2 max-h-60 overflow-y-auto">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-[var(--color-text-secondary)]">Search Results ({searchResults.length})</span>
            <button onClick={() => setSearchResults([])} className="text-xs text-red-400">Clear</button>
          </div>
          {searchResults.map(chunk => (
            <div key={chunk.id} className="p-2 bg-[var(--color-bg-tertiary)] rounded text-xs">
              <p className="line-clamp-2">{chunk.content}</p>
              <span className="text-[var(--color-text-secondary)]">{chunk.token_count} tokens</span>
            </div>
          ))}
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : docs.length === 0 ? (
          <div className="text-center py-12">
            <BookOpen size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)] mb-3">No documents uploaded yet</p>
            <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm">
              Upload Your First Document
            </button>
          </div>
        ) : (
          <div className="space-y-2">
            {docs.map(doc => (
              <div key={doc.id} className="flex items-center gap-3 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] hover:border-[var(--color-primary)] transition-colors">
                <FileText size={18} className="text-[var(--color-text-secondary)] shrink-0" />
                <button onClick={() => selectDoc(doc)} className="flex-1 text-left">
                  <div className="text-sm font-medium">{doc.title}</div>
                  <div className="text-xs text-[var(--color-text-secondary)]">
                    {doc.file_type} &middot; {doc.chunk_count} chunks &middot; {doc.status}
                    {doc.status === 'error' && <span className="text-red-400 ml-1">{doc.error_message}</span>}
                  </div>
                </button>
                <button onClick={() => handleDelete(doc.id)} className="text-red-400 hover:text-red-300 p-1">
                  <Trash2 size={16} />
                </button>
                <button onClick={() => selectDoc(doc)} className="text-[var(--color-text-secondary)]">
                  <ChevronRight size={16} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-[var(--color-bg)] rounded-xl border border-[var(--color-border)] p-6 w-full max-w-lg space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold">Add Document</h3>
              <button onClick={() => setShowCreate(false)}><X size={18} /></button>
            </div>
            <input value={title} onChange={e => setTitle(e.target.value)} placeholder="Document title"
              className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
            <div className="flex gap-3">
              <input value={fileName} onChange={e => setFileName(e.target.value)} placeholder="File name (optional)"
                className="flex-1 px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm" />
              <select value={fileType} onChange={e => setFileType(e.target.value)}
                className="w-32 px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm">
                <option value="text">Text</option>
                <option value="markdown">Markdown</option>
                <option value="code">Code</option>
                <option value="pdf">PDF</option>
              </select>
            </div>
            <textarea value={content} onChange={e => setContent(e.target.value)} placeholder="Paste document content..."
              rows={10} className="w-full px-3 py-2 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg text-sm font-mono" />
            <button onClick={handleCreate} className="w-full py-2 bg-[var(--color-primary)] text-white rounded-lg text-sm hover:opacity-90">
              Upload & Process
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

// ---- Versions Tab ----

function VersionsTab({ projectId }: { projectId: string }) {
  const [memories, setMemories] = useState<TeamMemory[]>([]);
  const [selectedMemory, setSelectedMemory] = useState<TeamMemory | null>(null);
  const [versions, setVersions] = useState<MemoryVersion[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadMemories();
  }, [projectId]);

  const loadMemories = async () => {
    setLoading(true);
    try {
      const res = await memoryApi.list(projectId);
      setMemories(res.data.memories || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  const selectMemory = async (mem: TeamMemory) => {
    setSelectedMemory(mem);
    try {
      const res = await knowledgeApi.listVersions(projectId, mem.id);
      setVersions(res.data.versions || []);
    } catch { /* empty */ }
  };

  const handleRollback = async (versionNumber: number) => {
    if (!selectedMemory) return;
    try {
      await knowledgeApi.rollback(projectId, selectedMemory.id, versionNumber);
      selectMemory(selectedMemory);
    } catch { /* empty */ }
  };

  if (selectedMemory) {
    return (
      <div className="flex flex-col h-full">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-border)] shrink-0">
          <button onClick={() => setSelectedMemory(null)} className="text-sm text-[var(--color-primary)]">&larr; Back</button>
          <div className="flex-1">
            <h3 className="font-semibold text-sm">{selectedMemory.category}/{selectedMemory.key}</h3>
            <span className="text-xs text-[var(--color-text-secondary)]">{versions.length} versions</span>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {versions.length === 0 ? (
            <p className="text-xs text-[var(--color-text-secondary)] text-center py-8">No version history available.</p>
          ) : versions.map(v => (
            <div key={v.id} className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-mono font-bold">v{v.version_number}</span>
                  <span className={`text-xs px-1.5 py-0.5 rounded ${v.change_type === 'create' ? 'bg-green-500/20 text-green-400' : v.change_type === 'rollback' ? 'bg-yellow-500/20 text-yellow-400' : 'bg-blue-500/20 text-blue-400'}`}>
                    {v.change_type}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-[var(--color-text-secondary)]">{new Date(v.created_at).toLocaleString()}</span>
                  <button onClick={() => handleRollback(v.version_number)} className="flex items-center gap-1 text-xs text-[var(--color-primary)] hover:underline" title="Rollback to this version">
                    <RotateCcw size={12} /> Restore
                  </button>
                </div>
              </div>
              {v.diff_summary && <p className="text-xs text-[var(--color-text-secondary)] mb-1">{v.diff_summary}</p>}
              <pre className="text-xs whitespace-pre-wrap text-[var(--color-text-secondary)] max-h-32 overflow-y-auto">{v.content.substring(0, 300)}{v.content.length > 300 ? '...' : ''}</pre>
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <h3 className="font-semibold text-sm">Memory Version History</h3>
        <p className="text-xs text-[var(--color-text-secondary)]">Select a memory entry to view its change history</p>
      </div>
      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : memories.length === 0 ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">No memories to show.</p>
        ) : (
          <div className="space-y-2">
            {memories.map(mem => (
              <button key={mem.id} onClick={() => selectMemory(mem)}
                className="w-full flex items-center gap-3 p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] hover:border-[var(--color-primary)] transition-colors text-left">
                <div className="flex-1">
                  <div className="text-sm font-medium">{mem.key}</div>
                  <div className="text-xs text-[var(--color-text-secondary)]">{mem.category} &middot; {new Date(mem.updated_at).toLocaleDateString()}</div>
                </div>
                <ChevronRight size={16} className="text-[var(--color-text-secondary)]" />
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// ---- Shared Tab ----

function SharedTab({ projectId }: { projectId: string }) {
  const [shared, setShared] = useState<SharedMemory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => { loadShared(); }, [projectId]);

  const loadShared = async () => {
    setLoading(true);
    try {
      const res = await knowledgeApi.getSharedForProject(projectId);
      setShared(res.data.shared_memories || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <h3 className="font-semibold text-sm">Shared Knowledge</h3>
        <p className="text-xs text-[var(--color-text-secondary)]">Memories shared across projects in your organization</p>
      </div>
      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : shared.length === 0 ? (
          <div className="text-center py-12">
            <Share2 size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)]">No shared memories available</p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">Share memories from the Memory tab to make them available here</p>
          </div>
        ) : (
          <div className="space-y-2">
            {shared.map(s => (
              <div key={s.id} className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                <div className="flex items-center justify-between mb-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{s.memory?.key || 'Unknown'}</span>
                    <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]">
                      {s.memory?.category}
                    </span>
                  </div>
                  <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)]">
                    {s.access_level}
                  </span>
                </div>
                <p className="text-xs text-[var(--color-text-secondary)] line-clamp-2">{s.memory?.content}</p>
                {s.project && (
                  <span className="text-xs text-[var(--color-text-secondary)] mt-1 block">From: {s.project.name}</span>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// ---- Extractions Tab ----

function ExtractionsTab({ projectId }: { projectId: string }) {
  const [extractions, setExtractions] = useState<AutoExtraction[]>([]);
  const [filter, setFilter] = useState<'pending' | 'accepted' | 'rejected' | ''>('pending');
  const [loading, setLoading] = useState(true);

  useEffect(() => { loadExtractions(); }, [projectId, filter]);

  const loadExtractions = async () => {
    setLoading(true);
    try {
      const res = await knowledgeApi.listExtractions(projectId, filter || undefined, 50);
      setExtractions(res.data.extractions || []);
    } catch { /* empty */ }
    setLoading(false);
  };

  const handleAccept = async (id: string, saveToMemory: boolean) => {
    try {
      await knowledgeApi.acceptExtraction(projectId, id, saveToMemory);
      loadExtractions();
    } catch { /* empty */ }
  };

  const handleReject = async (id: string) => {
    try {
      await knowledgeApi.rejectExtraction(projectId, id);
      loadExtractions();
    } catch { /* empty */ }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] shrink-0">
        <div>
          <h3 className="font-semibold text-sm">Auto-Extracted Knowledge</h3>
          <p className="text-xs text-[var(--color-text-secondary)]">Insights automatically identified from chat conversations</p>
        </div>
      </div>
      <div className="flex gap-2 px-4 py-2 border-b border-[var(--color-border)] shrink-0">
        {(['pending', 'accepted', 'rejected', ''] as const).map(f => (
          <button key={f} onClick={() => setFilter(f)}
            className={`px-3 py-1 text-xs rounded-lg ${filter === f ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg-secondary)] text-[var(--color-text-secondary)]'}`}>
            {f || 'All'}
          </button>
        ))}
      </div>
      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">Loading...</p>
        ) : extractions.length === 0 ? (
          <div className="text-center py-12">
            <Sparkles size={40} className="mx-auto text-[var(--color-text-secondary)] mb-3 opacity-50" />
            <p className="text-sm text-[var(--color-text-secondary)]">No extractions found</p>
          </div>
        ) : (
          <div className="space-y-2">
            {extractions.map(ext => (
              <div key={ext.id} className="p-3 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)]">
                <div className="flex items-center gap-2 mb-2">
                  <span className={`text-xs px-1.5 py-0.5 rounded ${extractionColors[ext.extraction_type] || 'bg-gray-500/20 text-gray-400'}`}>
                    {ext.extraction_type}
                  </span>
                  <span className="text-xs text-[var(--color-text-secondary)]">
                    {(ext.confidence * 100).toFixed(0)}% confidence
                  </span>
                  <span className="text-xs text-[var(--color-text-secondary)] ml-auto">
                    {new Date(ext.created_at).toLocaleString()}
                  </span>
                </div>
                <p className="text-sm mb-2">{ext.extracted_content}</p>
                {ext.accepted === null && (
                  <div className="flex gap-2">
                    <button onClick={() => handleAccept(ext.id, true)} className="flex items-center gap-1 px-2 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700">
                      <Check size={12} /> Accept & Save
                    </button>
                    <button onClick={() => handleAccept(ext.id, false)} className="flex items-center gap-1 px-2 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700">
                      <Check size={12} /> Accept
                    </button>
                    <button onClick={() => handleReject(ext.id)} className="flex items-center gap-1 px-2 py-1 text-xs bg-red-600/20 text-red-400 rounded hover:bg-red-600/30">
                      <XCircle size={12} /> Reject
                    </button>
                  </div>
                )}
                {ext.accepted === true && (
                  <span className="text-xs text-green-400">Accepted {ext.memory_id ? '& saved to memory' : ''}</span>
                )}
                {ext.accepted === false && (
                  <span className="text-xs text-red-400">Rejected</span>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
