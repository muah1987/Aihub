import { useState, useEffect } from 'react';
import { Sun, Moon, Monitor, Palette, Type, Volume2, VolumeX, Globe } from 'lucide-react';
import { platformApi, type UserPreference } from '../../api/platform';

const themes = [
  { value: 'light', label: 'Light', icon: Sun },
  { value: 'dark', label: 'Dark', icon: Moon },
  { value: 'system', label: 'System', icon: Monitor },
] as const;

const accentColors = ['blue', 'purple', 'green', 'orange', 'red', 'pink', 'teal', 'amber'];

export function PreferencesPanel() {
  const [prefs, setPrefs] = useState<UserPreference | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    platformApi.getPreferences()
      .then((res) => setPrefs(res.data.preferences))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const updatePref = async (updates: Partial<UserPreference>) => {
    setSaving(true);
    try {
      const res = await platformApi.updatePreferences(updates);
      setPrefs(res.data.preferences);
    } catch {
      // silent
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <div className="p-6 text-sm text-[var(--color-text-secondary)]">Loading preferences...</div>;
  }

  if (!prefs) return null;

  return (
    <div className="p-6 max-w-2xl space-y-8">
      <div>
        <h2 className="text-lg font-semibold mb-1">Preferences</h2>
        <p className="text-sm text-[var(--color-text-secondary)]">Customize your Aihub experience</p>
      </div>

      {/* Theme */}
      <section>
        <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
          <Palette size={16} /> Theme
        </h3>
        <div className="flex gap-2">
          {themes.map(({ value, label, icon: Icon }) => (
            <button
              key={value}
              onClick={() => updatePref({ theme: value })}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg border text-sm transition-colors ${
                prefs.theme === value
                  ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/10 text-[var(--color-primary)]'
                  : 'border-[var(--color-border)] hover:border-[var(--color-text-secondary)]'
              }`}
            >
              <Icon size={16} />
              {label}
            </button>
          ))}
        </div>
      </section>

      {/* Accent color */}
      <section>
        <h3 className="text-sm font-medium mb-3">Accent Color</h3>
        <div className="flex gap-2 flex-wrap">
          {accentColors.map((color) => (
            <button
              key={color}
              onClick={() => updatePref({ accent_color: color })}
              className={`w-8 h-8 rounded-full border-2 transition-transform hover:scale-110 ${
                prefs.accent_color === color ? 'border-[var(--color-text)] scale-110' : 'border-transparent'
              }`}
              style={{ backgroundColor: `var(--accent-${color}, ${color})` }}
              title={color}
            />
          ))}
        </div>
      </section>

      {/* Editor font size */}
      <section>
        <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
          <Type size={16} /> Editor Font Size
        </h3>
        <div className="flex items-center gap-3">
          <input
            type="range"
            min={10}
            max={24}
            value={prefs.editor_font_size}
            onChange={(e) => updatePref({ editor_font_size: Number(e.target.value) })}
            className="flex-1 accent-[var(--color-primary)]"
          />
          <span className="text-sm font-mono w-8 text-center">{prefs.editor_font_size}</span>
        </div>
      </section>

      {/* Layout */}
      <section>
        <h3 className="text-sm font-medium mb-3">Layout</h3>
        <div className="space-y-3">
          <label className="flex items-center justify-between cursor-pointer">
            <span className="text-sm">Compact mode</span>
            <input
              type="checkbox"
              checked={prefs.compact_mode}
              onChange={(e) => updatePref({ compact_mode: e.target.checked })}
              className="accent-[var(--color-primary)]"
            />
          </label>
          <label className="flex items-center justify-between cursor-pointer">
            <span className="text-sm">Sidebar collapsed by default</span>
            <input
              type="checkbox"
              checked={prefs.sidebar_collapsed}
              onChange={(e) => updatePref({ sidebar_collapsed: e.target.checked })}
              className="accent-[var(--color-primary)]"
            />
          </label>
        </div>
      </section>

      {/* Notifications */}
      <section>
        <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
          {prefs.notifications_sound ? <Volume2 size={16} /> : <VolumeX size={16} />}
          Notifications
        </h3>
        <label className="flex items-center justify-between cursor-pointer">
          <span className="text-sm">Notification sounds</span>
          <input
            type="checkbox"
            checked={prefs.notifications_sound}
            onChange={(e) => updatePref({ notifications_sound: e.target.checked })}
            className="accent-[var(--color-primary)]"
          />
        </label>
      </section>

      {/* Locale & Timezone */}
      <section>
        <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
          <Globe size={16} /> Regional
        </h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="text-xs text-[var(--color-text-secondary)] mb-1 block">Locale</label>
            <select
              value={prefs.locale}
              onChange={(e) => updatePref({ locale: e.target.value })}
              className="w-full bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm"
            >
              <option value="en">English</option>
              <option value="es">Español</option>
              <option value="fr">Français</option>
              <option value="de">Deutsch</option>
              <option value="ja">日本語</option>
              <option value="zh">中文</option>
            </select>
          </div>
          <div>
            <label className="text-xs text-[var(--color-text-secondary)] mb-1 block">Timezone</label>
            <select
              value={prefs.timezone}
              onChange={(e) => updatePref({ timezone: e.target.value })}
              className="w-full bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm"
            >
              <option value="UTC">UTC</option>
              <option value="America/New_York">Eastern (US)</option>
              <option value="America/Chicago">Central (US)</option>
              <option value="America/Denver">Mountain (US)</option>
              <option value="America/Los_Angeles">Pacific (US)</option>
              <option value="Europe/London">London</option>
              <option value="Europe/Berlin">Berlin</option>
              <option value="Asia/Tokyo">Tokyo</option>
              <option value="Asia/Shanghai">Shanghai</option>
            </select>
          </div>
        </div>
      </section>

      {saving && (
        <div className="text-xs text-[var(--color-text-secondary)]">Saving...</div>
      )}
    </div>
  );
}
