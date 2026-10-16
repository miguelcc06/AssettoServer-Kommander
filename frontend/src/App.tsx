import { useEffect, useState, useRef } from 'react'
import './index.css'
import { EventBuilderView } from './EventBuilderView'
import { LogsView } from './LogsView'
import LiveTiming from './LiveTiming'
import UsersView from './UsersView'
import { PluginsView } from './PluginsView'

interface CarMetric {
  model: string;
  skin: string;
  driverName: string;
  driverTeam: string;
  isConnected: boolean;
  bestLap?: number;
}

interface LiveMetrics {
  info: {
    name: string;
    track: string;
    maxclients: number;
    clients: number;
    ip: string;
    cport: number;
  };
  cars: CarMetric[];
}

function GeneralConfigView() {
  const [config, setConfig] = useState({
    serverName: '',
    serverPassword: '',
    adminPassword: '',
    httpPort: '',
    tcpPort: '',
    udpPort: '',
  })
  const [msg, setMsg] = useState({ type: '', text: '' })
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    fetch('/api/config/general')
      .then(res => res.json())
      .then(data => setConfig({
        serverName: data.serverName || '',
        serverPassword: data.serverPassword || '',
        adminPassword: data.adminPassword || '',
        httpPort: data.httpPort || '',
        tcpPort: data.tcpPort || '',
        udpPort: data.udpPort || '',
      }))
      .catch(err => setMsg({ type: 'error', text: 'Loading error: ' + err.message }))
  }, [])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setConfig({ ...config, [e.target.name]: e.target.value })
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setMsg({ type: '', text: '' })
    try {
      const res = await fetch('/api/config/general', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      })
      if (res.ok) {
        setMsg({ type: 'success', text: 'Configuration saved successfully.' })
      } else {
        setMsg({ type: 'error', text: 'Error saving configuration.' })
      }
    } catch (err: any) {
      setMsg({ type: 'error', text: 'Network error: ' + err.message })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="card form-card">
      <h2 style={{ marginTop: 0 }}>General Configuration</h2>
      
      {msg.text && (
        <div className={msg.type === 'success' ? 'msg-success' : 'msg-error'}>
          {msg.text}
        </div>
      )}

      <form onSubmit={handleSave}>
        <div className="form-group">
          <label>Server Name</label>
          <input type="text" name="serverName" value={config.serverName} onChange={handleChange} />
        </div>
        <div className="form-group">
          <label>Server Password (optional)</label>
          <input type="text" name="serverPassword" value={config.serverPassword} onChange={handleChange} />
        </div>
        <div className="form-group">
          <label>Admin Password</label>
          <input type="text" name="adminPassword" value={config.adminPassword} onChange={handleChange} />
        </div>
        
        <h3 style={{ margin: '24px 0 12px', fontSize: '1.1rem' }}>Network Ports</h3>
        <div className="form-group">
          <label>HTTP Port</label>
          <input type="number" name="httpPort" value={config.httpPort} onChange={handleChange} />
        </div>
        <div className="form-group">
          <label>TCP Port</label>
          <input type="number" name="tcpPort" value={config.tcpPort} onChange={handleChange} />
        </div>
        <div className="form-group">
          <label>UDP Port</label>
          <input type="number" name="udpPort" value={config.udpPort} onChange={handleChange} />
        </div>

        <button type="submit" className="btn-save" disabled={saving}>
          {saving ? 'Saving...' : 'Save'}
        </button>
      </form>
    </div>
  )
}

function EventConfigView() {
  const [config, setConfig] = useState({
    track: '',
    trackLayout: '',
    maxClients: '',
    practiceTime: '',
    qualifyTime: '',
    raceLaps: '',
  })
  const [msg, setMsg] = useState({ type: '', text: '' })
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    fetch('/api/config/event')
      .then(res => res.json())
      .then(data => setConfig({
        track: data.track || '',
        trackLayout: data.trackLayout || '',
        maxClients: data.maxClients || '',
        practiceTime: data.practiceTime || '',
        qualifyTime: data.qualifyTime || '',
        raceLaps: data.raceLaps || '',
      }))
      .catch(err => setMsg({ type: 'error', text: 'Loading error: ' + err.message }))
  }, [])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setConfig({ ...config, [e.target.name]: e.target.value })
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setMsg({ type: '', text: '' })
    try {
      const res = await fetch('/api/config/event', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      })
      if (res.ok) {
        setMsg({ type: 'success', text: 'Event saved successfully.' })
      } else {
        setMsg({ type: 'error', text: 'Error saving configuration.' })
      }
    } catch (err: any) {
      setMsg({ type: 'error', text: 'Network error: ' + err.message })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="card form-card">
      <h2 style={{ marginTop: 0 }}>Event Configuration</h2>
      
      {msg.text && (
        <div className={msg.type === 'success' ? 'msg-success' : 'msg-error'}>
          {msg.text}
        </div>
      )}

      <form onSubmit={handleSave}>
        <div className="form-group">
          <label>Track (e.g., ks_monza)</label>
          <input type="text" name="track" value={config.track} onChange={handleChange} />
        </div>
        <div className="form-group">
          <label>Track Layout (e.g., gp)</label>
          <input type="text" name="trackLayout" value={config.trackLayout} onChange={handleChange} />
        </div>
        <div className="form-group">
          <label>Max Clients</label>
          <input type="number" name="maxClients" value={config.maxClients} onChange={handleChange} />
        </div>
        
        <h3 style={{ margin: '24px 0 12px', fontSize: '1.1rem' }}>Sessions</h3>
        <div className="form-group">
          <label>Practice (Minutes)</label>
          <input type="number" name="practiceTime" value={config.practiceTime} onChange={handleChange} />
        </div>
        <div className="form-group">
          <label>Qualifying (Minutes)</label>
          <input type="number" name="qualifyTime" value={config.qualifyTime} onChange={handleChange} />
        </div>
        <div className="form-group">
          <label>Race (Laps)</label>
          <input type="number" name="raceLaps" value={config.raceLaps} onChange={handleChange} />
        </div>

        <button type="submit" className="btn-save" disabled={saving}>
          {saving ? 'Saving...' : 'Save'}
        </button>
      </form>
    </div>
  )
}


interface Car {
  id: string
  name: string
  brand: string
  description: string
  class: string
  specs: {
    bhp: string;
    torque: string;
    weight: string;
    topspeed: string;
  };
  skins?: string[];
}

interface TrackLayout {
  id: string
  name: string
  description: string
  length: string
  pitboxes: string
}

interface Track {
  id: string
  name: string
  description: string
  country: string
  length: string
  pitboxes: string
  layouts?: TrackLayout[]
}

interface InstalledContent {
  type: string
  id: string
  url: string
  version?: string
}

function PublicModeConfigView() {
  const [isPublic, setIsPublic] = useState(false)
  const [msg, setMsg] = useState({ type: '', text: '' })

  useEffect(() => {
    fetch('/api/settings?key=public_live_timing')
      .then(res => res.json())
      .then(data => {
        setIsPublic(data.value === 'true')
      })
  }, [])

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setMsg({ type: '', text: '' })
    try {
      const res = await fetch('/api/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: 'public_live_timing', value: isPublic ? 'true' : 'false' })
      })
      if (res.ok) {
        setMsg({ type: 'success', text: 'Setting saved successfully.' })
      } else {
        setMsg({ type: 'error', text: 'Error saving configuration.' })
      }
    } catch (err: any) {
      setMsg({ type: 'error', text: err.message })
    }
  }

  return (
    <div className="card form-card">
      <h2 style={{ marginTop: 0 }}>Public Live Timing Access</h2>
      <p style={{ color: 'var(--text-muted)', marginBottom: '20px', fontSize: '0.9rem' }}>
        Enabling this option allows anyone to view Live Timing without logging in, by navigating to the URL with <code>?live=true</code>.
      </p>

      {msg.text && (
        <div className={msg.type === 'success' ? 'msg-success' : 'msg-error'}>
          {msg.text}
        </div>
      )}

      <form onSubmit={handleSave}>
        <div className="form-group" style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <input 
            type="checkbox" 
            checked={isPublic} 
            onChange={(e) => setIsPublic(e.target.checked)} 
            id="publicLiveTimingToggle"
            style={{ width: '20px', height: '20px' }}
          />
          <label htmlFor="publicLiveTimingToggle" style={{ margin: 0, cursor: 'pointer' }}>Enable public mode for Live Timing</label>
        </div>

        <button type="submit" className="btn-save" style={{ marginTop: '12px' }}>
          Save
        </button>
      </form>
    </div>
  )
}

function SystemUpgradeView() {
  const [mode, setMode] = useState<'archive' | 'manual'>('archive')
  const [showManualModal, setShowManualModal] = useState(false)
  const [showPatreonModal, setShowPatreonModal] = useState(false)
  const [zipFile, setZipFile] = useState<File | null>(null)
  const [image, setImage] = useState('compujuckel/assettoserver:latest')
  const [patreonKey, setPatreonKey] = useState('')
  const [updateMsg, setUpdateMsg] = useState({ type: '', text: '' })
  const [patreonMsg, setPatreonMsg] = useState({ type: '', text: '' })
  const [upgrading, setUpgrading] = useState(false)

  const handleKeyUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        if (event.target?.result) {
          setPatreonKey((event.target.result as string).trim());
        }
      };
      reader.readAsText(file);
    }
  }
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [autoUpdateEnabled, setAutoUpdateEnabled] = useState(false)

  useEffect(() => {
    fetch('/api/settings?key=auto_update_enabled')
      .then(res => res.json())
      .then(data => {
        if (data.value === 'true') setAutoUpdateEnabled(true)
      })
      .catch(err => console.error(err))

    fetch('/api/server/upgrade')
      .then(res => res.json())
      .then(data => {
        
        if (data.patreonKey) setPatreonKey(data.patreonKey)
      })
      .catch(err => console.error(err))
  }, [])

  const handleToggleAutoUpdate = async (enabled: boolean) => {
    setAutoUpdateEnabled(enabled)
    try {
      await fetch('/api/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: 'auto_update_enabled', value: enabled ? 'true' : 'false' })
      })
    } catch (err) {
      console.error("Failed to save auto update setting", err)
    }
  }

  const handleCheckUpdate = async () => {
    setCheckingUpdate(true)
    setUpdateMsg({ type: '', text: '' })
    try {
      const res = await fetch('/api/upgrade/check', { method: 'POST' })
      const data = await res.json()
      if (res.ok) {
        setUpdateMsg({ type: data.updated ? 'success' : 'success', text: data.message })
      } else {
        setUpdateMsg({ type: 'error', text: data.error || 'Erreur lors de la vérification.' })
      }
    } catch (err: any) {
      setUpdateMsg({ type: 'error', text: 'Network error: ' + err.message })
    } finally {
      setCheckingUpdate(false)
    }
  }

  const handleUpgrade = async (e: React.FormEvent) => {
    e.preventDefault()
    setUpgrading(true)
    setPatreonMsg({ type: 'info', text: 'Configuration appliquée. Restarting...' })
    try {
      if (mode === 'archive') {
        if (!zipFile) {
          setPatreonMsg({ type: 'error', text: 'Veuillez sélectionner un fichier .zip Patreon.' })
          setUpgrading(false)
          return
        }
        const formData = new FormData()
        formData.append('zipfile', zipFile)
        formData.append('patreonKey', patreonKey)

        const res = await fetch('/api/server/upgrade/zip', {
          method: 'POST',
          body: formData
        })
        const data = await res.json()
        if (res.ok) {
          let attempts = 0
          const poll = setInterval(async () => {
            attempts++
            try {
              const statRes = await fetch('/api/server/status')
              const statData = await statRes.json()
              if (statData.status === 'Running') {
                clearInterval(poll)
                setPatreonMsg({ type: 'success', text: 'Archive Patreon installée et Server en ligne !' })
              }
            } catch(e) {}
            if (attempts > 30) {
              clearInterval(poll)
              setPatreonMsg({ type: 'error', text: 'Le serveur met du temps à répondre. Vérifiez les logs.' })
            }
          }, 2000)
        } else {
          setPatreonMsg({ type: 'error', text: data.message || 'Erreur lors de la mise à niveau.' })
        }
      } else {
        const res = await fetch('/api/server/upgrade', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ 
            isAutoMode: false,
            image, 
            patreonKey 
          })
        })
        const data = await res.json()
        if (res.ok) {
          let attempts = 0
          const poll = setInterval(async () => {
            attempts++
            try {
              const statRes = await fetch('/api/server/status')
              const statData = await statRes.json()
              if (statData.status === 'Running') {
                clearInterval(poll)
                setPatreonMsg({ type: 'success', text: 'Configuration sauvegardée et Server en ligne !' })
              }
            } catch(e) {}
            if (attempts > 15) {
              clearInterval(poll)
              setPatreonMsg({ type: 'error', text: 'Le serveur met du temps à répondre. Vérifiez les logs.' })
            }
          }, 2000)
        } else {
          setPatreonMsg({ type: 'error', text: data.message || 'Erreur lors de la mise à niveau.' })
        }
      }
    } catch (err: any) {
      setPatreonMsg({ type: 'error', text: 'Network error: ' + err.message })
    } finally {
      setUpgrading(false)
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      
      {/* -------------------- Server Updates -------------------- */}
      <div className="card form-card">
        <h2 style={{ marginTop: 0 }}>Server Updates</h2>
        
        <p style={{ color: 'var(--text-color)', marginBottom: '15px' }}>
          <strong>Image Version:</strong> <code>{image}</code>
          <br/>
          <a href="https://github.com/compujuckel/AssettoServer/releases" target="_blank" rel="noreferrer" style={{ color: '#8a2be2', textDecoration: 'none', fontSize: '0.9rem', display: 'inline-block', marginTop: '5px' }}>
            🔗 View official Changelogs
          </a>
        </p>

        {updateMsg.text && (
          <div className={updateMsg.type === 'success' ? 'msg-success' : 'msg-error'}>
            {updateMsg.text}
          </div>
        )}

        
      </div>

      {/* -------------------- Patreon / GitHub Configuration -------------------- */}
      <div className="card form-card">
        <h2 style={{ marginTop: 0 }}>Patreon / GitHub Configuration</h2>
      
      <div style={{ display: 'flex', gap: '15px', marginBottom: '20px' }}>
        <button 
          onClick={() => setMode('archive')}
          style={{
            flex: 1, padding: '10px', borderRadius: '6px', cursor: 'pointer',
            backgroundColor: mode === 'archive' ? '#8a2be2' : 'var(--sidebar-bg)',
            color: 'white', border: 'none', fontWeight: 'bold'
          }}
        >
          Archive Mode (.zip)
        </button>
        <button 
          onClick={() => setMode('manual')}
          style={{
            flex: 1, padding: '10px', borderRadius: '6px', cursor: 'pointer',
            backgroundColor: mode === 'manual' ? '#8a2be2' : 'var(--sidebar-bg)',
            color: 'white', border: 'none', fontWeight: 'bold'
          }}
        >
          Manual Mode
        </button>
      </div>

      <p style={{ color: 'var(--text-muted)', marginBottom: '20px', fontSize: '0.9rem' }}>
        {mode === 'archive' 
          ? "Import the Patreon version .zip archive. The server will extract the files and build the Docker image with your plugins." 
          : "Manually enter the Premium version Docker image name if you have already downloaded or built it locally."}
      </p>

      {patreonMsg.text && (
        <div 
          className={patreonMsg.type === 'success' ? 'msg-success' : (patreonMsg.type === 'error' ? 'msg-error' : '')}
          style={patreonMsg.type === 'info' ? { backgroundColor: 'rgba(56, 189, 248, 0.1)', color: '#38bdf8', padding: '12px', borderRadius: '6px', marginBottom: '16px', border: '1px solid rgba(56, 189, 248, 0.2)' } : {}}
        >
          {patreonMsg.text}
        </div>
      )}

      <form onSubmit={handleUpgrade}>
        {mode === 'archive' ? (
          <>
            <div className="form-group">
              <label>Patreon Archive (.zip)</label>
              <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '-5px', marginBottom: '10px' }}>
                Download the file <code>assetto-server-patreon-*-linux-x64.zip</code> on <a href="https://patreon.assettoserver.org/key" target="_blank" rel="noreferrer" style={{ color: '#8a2be2' }}>the Patreon page</a>, then select it here.
              </p>
              <input 
                type="file" 
                accept=".zip" 
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) setZipFile(file);
                }}
                style={{ padding: '10px', background: 'rgba(0,0,0,0.2)', borderRadius: '6px', width: '100%', cursor: 'pointer' }}
              />
              {zipFile && <span style={{ color: '#4ade80', fontSize: '0.9rem', display: 'block', marginTop: '5px' }}>Selected file: {zipFile.name}</span>}
            </div>
          </>
        ) : (
          <>
            <div className="form-group">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <label>AssettoServer Docker Image</label>
                <a href="#" onClick={(e) => { e.preventDefault(); setShowManualModal(true); }} style={{ fontSize: '0.85rem', color: '#8a2be2' }}>How to do it manually?</a>
              </div>
              <input 
                type="text" 
                value={image} 
                onChange={(e) => setImage(e.target.value)} 
                placeholder="ex: assettoserver-patreon:latest"
              />
            </div>
          </>
        )}

        <div className="form-group">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <label>Patreon Key (PatreonHubPlugin)</label>
            <div style={{ display: 'flex', gap: '10px' }}>
              <label style={{ cursor: 'pointer', fontSize: '0.85rem', color: '#8a2be2', display: 'flex', alignItems: 'center', gap: '5px' }}>
                <input type="file" accept=".txt,.key" style={{ display: 'none' }} onChange={handleKeyUpload} />
                📂 Import a file
              </label>
              <a href="#" onClick={(e) => { e.preventDefault(); setShowPatreonModal(true); }} style={{ fontSize: '0.85rem', color: '#8a2be2' }}>Where to find my Patreon Key?</a>
            </div>
          </div>
          <input 
            type="password" 
            value={patreonKey} 
            onChange={(e) => setPatreonKey(e.target.value)} 
            placeholder="Paste your Patreon key here, or import the file..."
          />
        </div>

        <button type="submit" className="btn-save" disabled={upgrading || checkingUpdate} style={{ width: '100%', backgroundColor: '#8a2be2' }}>
          {upgrading ? 'Configuring...' : 'Save & Apply Config'}
        </button>
      </form>

      {showManualModal && (
        <div className="modal-overlay" onClick={() => setShowManualModal(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()} style={{ maxWidth: '600px' }}>
            <button className="modal-close" onClick={() => setShowManualModal(false)}>×</button>
            <h2 style={{ marginTop: 0 }}>How to use Manual mode?</h2>
            <p style={{ color: 'var(--text-muted)' }}>
              If you do not want or cannot use the automatic GitHub connection, here is how to proceed:
            </p>
            <ol style={{ color: 'var(--text-color)', lineHeight: '1.6', paddingLeft: '20px' }}>
              <li>Go to the Patreon website and download the <strong>.zip</strong> file of the latest version (e.g. v0.0.39).</li>
              <li>Upload this `.zip` to your server (for example via SFTP or FTP).</li>
              <li>Unzip the `.zip` in a folder.</li>
              <li>In this folder, open a terminal and run the Docker command to build the image:<br/>
                <code style={{ background: 'var(--main-bg)', border: '1px solid var(--border-color)', padding: '4px 8px', borderRadius: '4px', display: 'inline-block', margin: '8px 0' }}>docker build -t assettoserver-patreon:latest .</code>
              </li>
              <li>Return to this Web panel, and enter <strong>assettoserver-patreon:latest</strong> in the Docker Image field.</li>
              <li>Enter your Patreon key, then click <strong>Save & Apply Config</strong>.</li>
            </ol>
          </div>
        </div>
      )}

      

      {showPatreonModal && (
        <div className="modal-overlay" onClick={() => setShowPatreonModal(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()} style={{ maxWidth: '600px' }}>
            <button className="modal-close" onClick={() => setShowPatreonModal(false)}>×</button>
            <h2 style={{ marginTop: 0 }}>Where to find my Patreon Key?</h2>
            <p style={{ color: 'var(--text-muted)' }}>
              The key serves as a license to unlock paid features on your server.
            </p>
            <ol style={{ paddingLeft: '20px', lineHeight: '1.6' }}>
              <li>Go to the <a href="https://www.patreon.com/settings/apps" target="_blank" rel="noreferrer" style={{ color: '#8a2be2' }}>Patreon</a> website and ensure your Discord account is connected.</li>
              <li>Join the AssettoServer Discord and verify you have the <strong>Contributor Tier 2</strong> (or higher) role.</li>
              <li>Visit the official authentication site: <a href="https://patreon.assettoserver.org/connect" target="_blank" rel="noreferrer" style={{ color: '#8a2be2' }}>patreon.assettoserver.org</a></li>
              <li>Login with Discord on this site.</li>
              <li>Click on the big red button <strong>Download Key</strong>.</li>
              <li>You can then either <strong>copy/paste</strong> the file contents into the field, or use the <strong>Import a file</strong>.</li>
            </ol>
          </div>
        </div>
      )}
    </div>
    </div>
  )
}



function ContentConfigView() {
  const [contentTab, setContentTab] = useState<'cars' | 'tracks' | 'upload'>('cars')
  
  const [msg, setMsg] = useState({ type: '', text: '' })
  const [uploading, setUploading] = useState(false)
  const [preparing, setPreparing] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [isDragActive, setIsDragActive] = useState(false)
  const [installed, setInstalled] = useState<InstalledContent[]>([])
  const archiveInputRef = useRef<HTMLInputElement>(null)
  const folderInputRef = useRef<HTMLInputElement>(null)
  
  const [cars, setCars] = useState<Car[]>([])
  const [tracks, setTracks] = useState<Track[]>([])
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedCar, setSelectedCar] = useState<Car | null>(null)
  const [selectedSkin, setSelectedSkin] = useState<string>('')
  const [selectedTrack, setSelectedTrack] = useState<Track | null>(null)

  useEffect(() => {
    const node = folderInputRef.current
    if (!node) return
    node.setAttribute('webkitdirectory', '')
    node.setAttribute('directory', '')
  }, [])

  useEffect(() => {
    if (contentTab === 'cars') {
      fetch('/api/content/cars')
        .then(res => res.json())
        .then(data => setCars(Array.isArray(data) ? data : []))
        .catch(err => console.error("Error fetching cars:", err))
    } else if (contentTab === 'tracks') {
      fetch('/api/content/tracks')
        .then(res => res.json())
        .then(data => setTracks(Array.isArray(data) ? data : []))
        .catch(err => console.error("Error fetching tracks:", err))
    }
  }, [contentTab])

  const filteredCars = cars.filter(c => 
    (c.name || c.id).toLowerCase().includes(searchTerm.toLowerCase()) || 
    (c.brand || '').toLowerCase().includes(searchTerm.toLowerCase())
  );

  const filteredTracks = tracks.filter(t => 
    (t.name || t.id).toLowerCase().includes(searchTerm.toLowerCase()) || 
    (t.country || '').toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragActive(true)
  }

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragActive(false)
  }

  const uploadFiles = (formData: FormData) => {
    setPreparing(false)
    setUploading(true)
    setUploadProgress(0)
    setInstalled([])
    setMsg({ type: 'info', text: 'Uploading (0%)...' })
    
    const xhr = new XMLHttpRequest()
    xhr.open('POST', '/api/content/upload', true)
    
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        const percent = Math.round((e.loaded / e.total) * 100)
        setUploadProgress(percent)
        if (percent >= 100) {
          setMsg({ type: 'info', text: 'Upload complete. Detecting, installing and packaging for Content Manager...' })
        } else {
          setMsg({ type: 'info', text: `Uploading (${percent}%)...` })
        }
      }
    }
    
    xhr.onload = () => {
      setUploading(false)
      setUploadProgress(0)
      if (archiveInputRef.current) archiveInputRef.current.value = ''
      if (folderInputRef.current) folderInputRef.current.value = ''
      let data: any = {}
      try { data = JSON.parse(xhr.responseText) } catch (e) {}
      if (xhr.status >= 200 && xhr.status < 300) {
        setInstalled(Array.isArray(data.installed) ? data.installed : [])
        setMsg({ type: 'success', text: data.message || 'Upload successful!' })
      } else {
        setMsg({ type: 'error', text: data.message || 'Error during upload' })
      }
    }
    
    xhr.onerror = () => {
      setUploading(false)
      setMsg({ type: 'error', text: 'Network error during upload.' })
    }
    
    xhr.send(formData)
  }

  const isArchiveName = (name: string) => {
    const lower = name.toLowerCase()
    return lower.endsWith('.zip') || lower.endsWith('.rar') || lower.endsWith('.7z')
  }

  const appendContentFile = (formData: FormData, file: File, relativePath: string) => {
    const rel = (relativePath || file.name).replace(/\\/g, '/')
    const base = rel.split('/').pop() || file.name
    // Go's multipart parser keeps only the base filename, so the relative
    // path travels in a parallel paths[] field.
    if (isArchiveName(base)) {
      formData.append('file', file, base)
      return
    }
    formData.append('paths[]', rel)
    formData.append('files[]', file, base)
  }

  const appendSelectedFiles = (files: FileList | File[], formData: FormData) => {
    let count = 0
    for (const file of Array.from(files)) {
      const relative = file.webkitRelativePath || file.name
      appendContentFile(formData, file, relative)
      count++
    }
    return count
  }

  const traverseDirectory = async (entry: FileSystemEntry, path: string, formData: FormData) => {
    if (entry.isFile) {
      const fileEntry = entry as FileSystemFileEntry
      return new Promise<void>((resolve) => {
        fileEntry.file((file) => {
          appendContentFile(formData, file, path + file.name)
          resolve()
        })
      })
    } else if (entry.isDirectory) {
      const dirEntry = entry as FileSystemDirectoryEntry
      const dirReader = dirEntry.createReader()
      
      const readEntries = () => {
        return new Promise<FileSystemEntry[]>((resolve) => {
          dirReader.readEntries((entries) => resolve(entries))
        })
      }

      let allEntries: FileSystemEntry[] = []
      let entries = await readEntries()
      while (entries.length > 0) {
        allEntries.push(...entries)
        entries = await readEntries()
      }

      for (const e of allEntries) {
        await traverseDirectory(e, path + entry.name + '/', formData)
      }
    }
  }

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragActive(false)
    if (uploading || preparing) return
    
    if (e.dataTransfer.items) {
      const formData = new FormData()
      let hasContent = false
      setPreparing(true)
      setMsg({ type: 'info', text: 'Reading dropped files...' })

      try {
        for (let i = 0; i < e.dataTransfer.items.length; i++) {
          const item = e.dataTransfer.items[i]
          if (item.kind === 'file') {
            const entry = item.webkitGetAsEntry()
            if (entry) {
              if (entry.isDirectory) {
                await traverseDirectory(entry, '', formData)
                hasContent = true
              } else {
                const file = item.getAsFile()
                if (file) {
                  appendSelectedFiles([file], formData)
                  hasContent = true
                }
              }
            }
          }
        }
      } catch (err: any) {
        setPreparing(false)
        setMsg({ type: 'error', text: 'Could not read the dropped folder: ' + err.message })
        return
      }

      if (hasContent) {
        uploadFiles(formData)
      } else {
        setPreparing(false)
        setMsg({ type: 'error', text: 'Drop a .zip, .rar, .7z archive or a car/track folder.' })
      }
    }
  }

  const handleBrowse = (files: FileList | null) => {
    if (!files || files.length === 0 || uploading || preparing) return
    const formData = new FormData()
    const count = appendSelectedFiles(files, formData)
    if (count === 0) {
      setMsg({ type: 'error', text: 'No files selected.' })
      return
    }
    uploadFiles(formData)
  }

  return (
    <div className="card content-card" style={{ padding: '0', display: 'flex', flexDirection: 'column' }}>
      
      {/* Sub Tabs */}
      <div className="content-tabs">
        <button className={contentTab === 'cars' ? 'active' : ''} onClick={() => setContentTab('cars')}>
          🚗 Cars
        </button>
        <button className={contentTab === 'tracks' ? 'active' : ''} onClick={() => setContentTab('tracks')}>
          🛣 Tracks
        </button>
        <button className={contentTab === 'upload' ? 'active' : ''} onClick={() => setContentTab('upload')}>
          📤 Upload
        </button>
        {contentTab !== 'upload' && (
            <div style={{ marginLeft: 'auto', padding: '10px' }}>
                <input 
                  type="search" 
                  placeholder="Search..." 
                  value={searchTerm}
                  onChange={e => setSearchTerm(e.target.value)}
                  style={{ padding: '8px', borderRadius: '4px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-color)', color: 'white' }}
                />
            </div>
        )}
      </div>

      <div style={{ padding: '30px' }}>
        {contentTab === 'upload' && (
          <>
            <h2 style={{ marginTop: 0 }}>Upload cars and tracks</h2>
            <p style={{ color: 'var(--text-muted)', marginBottom: '20px', fontSize: '0.9rem' }}>
              Drop a <strong>.zip</strong>, <strong>.rar</strong>, <strong>.7z</strong> or a car/track folder.
              Kommander detects the content from <code>ui/ui_car.json</code>, <code>data.acd</code> and <code>ui/ui_track.json</code>,
              installs it, and publishes a Content Manager package (<code>content/cars/&lt;id&gt;</code> or <code>content/tracks/&lt;id&gt;</code>).
            </p>

            {msg.text && (
              <div 
                className={msg.type === 'success' ? 'msg-success' : (msg.type === 'error' ? 'msg-error' : '')}
                style={msg.type === 'info' ? { backgroundColor: 'rgba(56, 189, 248, 0.1)', color: '#38bdf8', padding: '12px', borderRadius: '6px', marginBottom: '16px', border: '1px solid rgba(56, 189, 248, 0.2)' } : {}}
              >
                {msg.text}
              </div>
            )}

            {(uploading || preparing) && (
              <div
                className="upload-progress"
                role="progressbar"
                aria-valuemin={0}
                aria-valuemax={100}
                aria-valuenow={uploading && uploadProgress < 100 ? uploadProgress : undefined}
                aria-label="Upload progress"
              >
                <div
                  className={`upload-progress-bar ${preparing || uploadProgress >= 100 ? 'indeterminate' : ''}`}
                  style={preparing || uploadProgress >= 100 ? undefined : { width: `${Math.max(uploadProgress, 3)}%` }}
                />
              </div>
            )}

            <div 
              className={`dropzone ${isDragActive ? 'active' : ''} ${uploading || preparing ? 'uploading' : ''}`}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
            >
              {uploading || preparing ? (
                <div className="dropzone-content">
                  <div className="spinner"></div>
                  <p>
                    {preparing
                      ? 'Reading folder...'
                      : uploadProgress >= 100
                        ? 'Packaging for Content Manager...'
                        : `Uploading (${uploadProgress}%)`}
                  </p>
                </div>
              ) : (
                <div className="dropzone-content">
                  <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ marginBottom: '12px', color: 'var(--primary-color)' }}>
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                    <polyline points="17 8 12 3 7 8"></polyline>
                    <line x1="12" y1="3" x2="12" y2="15"></line>
                  </svg>
                  <p>Drag a car or track here</p>
                  <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', marginTop: '6px' }}>.zip, .rar, .7z or a folder</p>
                  <div className="upload-actions">
                    <button type="button" className="btn-save" onClick={() => archiveInputRef.current?.click()}>
                      Choose archive
                    </button>
                    <button type="button" className="btn-save upload-secondary" onClick={() => folderInputRef.current?.click()}>
                      Choose folder
                    </button>
                  </div>
                </div>
              )}
            </div>
            <input
              ref={archiveInputRef}
              type="file"
              accept=".zip,.rar,.7z,application/zip,application/x-7z-compressed,application/vnd.rar"
              multiple
              hidden
              onChange={(e) => handleBrowse(e.target.files)}
            />
            <input
              ref={folderInputRef}
              type="file"
              multiple
              hidden
              onChange={(e) => handleBrowse(e.target.files)}
            />

            {installed.length > 0 && (
              <div className="installed-list">
                <h3>Ready for Content Manager</h3>
                {installed.map((item) => (
                  <div className="installed-item" key={`${item.type}-${item.id}`}>
                    <span className={`installed-badge ${item.type === 'track' ? 'track' : 'car'}`}>
                      {item.type === 'track' ? 'Track' : 'Car'}
                    </span>
                    <div className="installed-meta">
                      <strong>{item.id}</strong>
                      {item.version ? <span>version {item.version}</span> : null}
                      <a href={item.url} target="_blank" rel="noreferrer">{item.url}</a>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}

        {contentTab === 'cars' && (
          <>
            <h2 style={{ marginTop: 0 }}>Vehicles ({filteredCars.length})</h2>
            <div className="grid-container">
              {filteredCars.map(c => (
                <div className="card-item" key={c.id} onClick={() => window.open(`/?car=${encodeURIComponent(c.id)}`, '_blank')}>
                  <div className="car-preview" style={{ backgroundImage: `url('/api/content/cars/image?id=${encodeURIComponent(c.id)}&v=2')` }}></div>
                  <h3>{c.name || c.id}</h3>
                  <p className="card-meta">{c.brand}</p>
                </div>
              ))}
            </div>
          </>
        )}

        {contentTab === 'tracks' && (
          <>
            <h2 style={{ marginTop: 0 }}>Tracks ({filteredTracks.length})</h2>
            <div className="grid-container">
              {filteredTracks.map(t => (
                <div className="card-item" key={t.id} onClick={() => window.open(`/?track=${encodeURIComponent(t.id)}`, '_blank')}>
                  <div className="track-preview" style={{ backgroundImage: `url('/api/content/tracks/image?id=${encodeURIComponent(t.id)}&v=1')` }}></div>
                  <h3>{t.name || t.id}</h3>
                  <p className="card-meta">{t.country}</p>
                  <p className="card-meta">{t.layouts?.length || 1} layouts</p>
                </div>
              ))}
            </div>
          </>
        )}
      </div>

      {/* Modals */}
      {selectedCar && (
        <div className="modal-overlay" onClick={() => setSelectedCar(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setSelectedCar(null)}>×</button>
            <h2>{selectedCar.name || selectedCar.id}</h2>
            <p style={{ color: 'var(--primary-color)', fontWeight: 'bold' }}>{selectedCar.brand} | {selectedCar.class}</p>
            <div className="modal-specs">
              <div><strong>BHP:</strong> {selectedCar.specs?.bhp || 'N/A'}</div>
              <div><strong>Torque:</strong> {selectedCar.specs?.torque || 'N/A'}</div>
              <div><strong>Weight:</strong> {selectedCar.specs?.weight || 'N/A'}</div>
              <div><strong>Top Speed:</strong> {selectedCar.specs?.topspeed || 'N/A'}</div>
            </div>
            <p className="modal-desc">{selectedCar.description}</p>
          </div>
        </div>
      )}

      {selectedTrack && (
        <div className="modal-overlay" onClick={() => setSelectedTrack(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setSelectedTrack(null)}>×</button>
            <h2>{selectedTrack.name || selectedTrack.id}</h2>
            <p style={{ color: 'var(--primary-color)', fontWeight: 'bold' }}>{selectedTrack.country}</p>
            <p className="modal-desc">{selectedTrack.description}</p>
            <h3>Layouts</h3>
            <div className="modal-list">
              {selectedTrack.layouts && selectedTrack.layouts.length > 0 ? (
                selectedTrack.layouts.map((layout, i) => (
                  <div key={i} className="modal-list-item">
                    <strong>{layout.name || layout.id}</strong> - {layout.length || '?'}m
                  </div>
                ))
              ) : (
                <div className="modal-list-item">
                  <strong>Default (Standard)</strong> - {selectedTrack.length || '?'}m
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function StandaloneCarView({ carId }: { carId: string }) {
  const [car, setCar] = useState<Car | null>(null);
  const [selectedSkin, setSelectedSkin] = useState<string>('');

  useEffect(() => {
    fetch('/api/content/cars')
      .then(res => res.json())
      .then((data: Car[]) => {
        const found = data.find(c => c.id === carId);
        if (found) {
          setCar(found);
          setSelectedSkin(found.skins && found.skins.length > 0 ? found.skins[0] : '');
        }
      });
  }, [carId]);

  if (!car) return <div style={{ color: 'var(--text-color)', padding: '20px', textAlign: 'center', marginTop: '50px' }}>Loading...</div>;

  return (
    <div style={{ minHeight: '100vh', backgroundColor: 'var(--bg-color)', padding: '40px', color: 'var(--text-color)' }}>
      <div style={{ maxWidth: '1200px', margin: '0 auto' }}>
        <button onClick={() => window.close()} style={{ background: 'var(--sidebar-bg)', border: '1px solid var(--border-color)', color: 'var(--text-color)', padding: '10px 20px', borderRadius: '6px', cursor: 'pointer', marginBottom: '30px', fontWeight: 'bold' }}>
          &larr; Close the tab
        </button>
        
        <h1 style={{ marginTop: 0, marginBottom: '10px', fontSize: '2.5rem' }}>{car.name || car.id}</h1>
        <p style={{ color: 'var(--primary-color)', fontSize: '1.2rem', marginBottom: '40px', fontWeight: 'bold' }}>{car.brand} <span style={{ color: 'var(--text-muted)' }}>• {car.class}</span></p>
        
        <div style={{ display: 'flex', flexDirection: 'column', gap: '40px' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
            <img 
              src={selectedSkin 
                ? `/api/content/cars/skin_image?id=${encodeURIComponent(car.id)}&skin=${encodeURIComponent(selectedSkin)}` 
                : `/api/content/cars/image?id=${encodeURIComponent(car.id)}`}
              alt={car.name || car.id}
              style={{ 
                width: '100%',
                height: 'auto',
                borderRadius: '12px',
                display: 'block'
              }}
            />
            
            {car.skins && car.skins.length > 0 && (
              <div style={{ marginTop: '20px' }}>
                <h3 style={{ marginBottom: '15px' }}>Available skins ({car.skins.length})</h3>
                <div className="skin-slider" style={{ display: 'flex', overflowX: 'auto', gap: '15px', paddingBottom: '15px' }}>
                  {car.skins.map(skin => (
                    <div 
                      key={skin}
                      onClick={() => setSelectedSkin(skin)}
                      title={skin}
                      style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '8px', cursor: 'pointer' }}
                    >
                      <div 
                        className={`skin-item ${selectedSkin === skin ? 'active' : ''}`}
                        style={{ 
                          minWidth: '160px', 
                          height: '100px', 
                          backgroundImage: `url('/api/content/cars/skin_image?id=${encodeURIComponent(car.id)}&skin=${encodeURIComponent(skin)}')` 
                        }}
                      ></div>
                      <span style={{ 
                        fontSize: '0.85rem', 
                        color: selectedSkin === skin ? 'var(--primary-color)' : 'var(--text-muted)', 
                        textAlign: 'center', 
                        maxWidth: '160px', 
                        overflow: 'hidden', 
                        textOverflow: 'ellipsis', 
                        whiteSpace: 'nowrap',
                        fontWeight: selectedSkin === skin ? 'bold' : 'normal'
                      }}>
                        {skin}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
          
          <div style={{ flex: '1 1 300px', backgroundColor: 'var(--sidebar-bg)', padding: '30px', borderRadius: '12px', height: 'fit-content', border: '1px solid var(--border-color)' }}>
            <h2 style={{ marginTop: 0, borderBottom: '2px solid var(--primary-color)', paddingBottom: '10px' }}>Specifications</h2>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '15px', marginTop: '20px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid rgba(255,255,255,0.05)', paddingBottom: '10px' }}>
                <span style={{ color: 'var(--text-muted)' }}>Power</span>
                <strong>{car.specs?.bhp || '?'}</strong>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid rgba(255,255,255,0.05)', paddingBottom: '10px' }}>
                <span style={{ color: 'var(--text-muted)' }}>Torque</span>
                <strong>{car.specs?.torque || '?'}</strong>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid rgba(255,255,255,0.05)', paddingBottom: '10px' }}>
                <span style={{ color: 'var(--text-muted)' }}>Weight</span>
                <strong>{car.specs?.weight || '?'}</strong>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: 'var(--text-muted)' }}>Top Speed</span>
                <strong>{car.specs?.topspeed || '?'}</strong>
              </div>
            </div>
            
            {car.description && (
              <div style={{ marginTop: '40px' }}>
                <h2>Description</h2>
                <p style={{ color: 'var(--text-muted)', lineHeight: '1.6', fontSize: '0.95rem' }}>{car.description}</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function StandaloneTrackView({ trackId }: { trackId: string }) {
  const [track, setTrack] = useState<Track | null>(null);
  const [selectedLayout, setSelectedLayout] = useState<string>('');
  const [uploading, setUploading] = useState(false);
  const [uploadMsg, setUploadMsg] = useState({ text: '', type: '' });
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    fetch('/api/content/tracks')
      .then(res => res.json())
      .then((data: Track[]) => {
        const found = data.find(t => t.id === trackId);
        if (found) {
          setTrack(found);
          if (found.layouts && found.layouts.length > 0) {
            setSelectedLayout(found.layouts[0].id);
          }
        }
      });
  }, [trackId]);

  if (!track) return <div style={{ color: 'var(--text-color)', padding: '20px', textAlign: 'center', marginTop: '50px' }}>Loading...</div>;

  return (
    <div style={{ minHeight: '100vh', backgroundColor: 'var(--bg-color)', padding: '40px', color: 'var(--text-color)' }}>
      <div style={{ maxWidth: '1200px', margin: '0 auto' }}>
        <button onClick={() => window.close()} style={{ background: 'var(--sidebar-bg)', border: '1px solid var(--border-color)', color: 'var(--text-color)', padding: '10px 20px', borderRadius: '6px', cursor: 'pointer', marginBottom: '30px', fontWeight: 'bold' }}>
          &larr; Close the tab
        </button>
        
        <h1 style={{ marginTop: 0, marginBottom: '10px', fontSize: '2.5rem' }}>
          {track.layouts?.find(l => l.id === selectedLayout)?.name || track.name || track.id}
        </h1>
        <p style={{ color: 'var(--primary-color)', fontSize: '1.2rem', marginBottom: '40px', fontWeight: 'bold' }}>
          {track.layouts?.find(l => l.id === selectedLayout)?.country || track.country}
        </p>
        
        <div style={{ display: 'flex', flexDirection: 'column', gap: '40px' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
            <img 
              src={selectedLayout 
                ? `/api/content/tracks/layout_image?id=${encodeURIComponent(track.id)}&layout=${encodeURIComponent(selectedLayout)}`
                : `/api/content/tracks/image?id=${encodeURIComponent(track.id)}`}
              alt={track.name || track.id}
              style={{ 
                width: '100%',
                height: 'auto',
                borderRadius: '12px',
                display: 'block'
              }}
            />

            {track.layouts && track.layouts.length > 0 && (
              <div style={{ marginTop: '20px' }}>
                <h3 style={{ marginBottom: '15px' }}>Available Layouts ({track.layouts.length})</h3>
                <div className="skin-slider" style={{ display: 'flex', overflowX: 'auto', gap: '15px', paddingBottom: '15px' }}>
                  {track.layouts.map(layout => (
                    <div 
                      key={layout.id}
                      onClick={() => setSelectedLayout(layout.id)}
                      title={layout.name || layout.id}
                      style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '8px', cursor: 'pointer' }}
                    >
                      <div 
                        className={`skin-item ${selectedLayout === layout.id ? 'active' : ''}`}
                        style={{ 
                          minWidth: '160px', 
                          height: '100px', 
                          backgroundImage: `url('/api/content/tracks/layout_image?id=${encodeURIComponent(track.id)}&layout=${encodeURIComponent(layout.id)}')`,
                          backgroundSize: 'cover',
                          backgroundPosition: 'center',
                          borderRadius: '8px'
                        }}
                      ></div>
                      <span style={{ 
                        fontSize: '0.85rem', 
                        color: selectedLayout === layout.id ? 'var(--primary-color)' : 'var(--text-muted)', 
                        textAlign: 'center', 
                        maxWidth: '160px', 
                        overflow: 'hidden', 
                        textOverflow: 'ellipsis', 
                        whiteSpace: 'nowrap',
                        fontWeight: selectedLayout === layout.id ? 'bold' : 'normal'
                      }}>
                        {layout.name || layout.id}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
          
          <div style={{ backgroundColor: 'var(--sidebar-bg)', padding: '30px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
            <h2 style={{ marginTop: 0, borderBottom: '2px solid var(--primary-color)', paddingBottom: '10px' }}>Information</h2>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '15px', marginTop: '20px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid rgba(255,255,255,0.05)', paddingBottom: '10px' }}>
                <span style={{ color: 'var(--text-muted)' }}>Length</span>
                <strong>{track.layouts?.find(l => l.id === selectedLayout)?.length || track.length || '?'}</strong>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid rgba(255,255,255,0.05)', paddingBottom: '10px' }}>
                <span style={{ color: 'var(--text-muted)' }}>Pitboxes</span>
                <strong>{track.layouts?.find(l => l.id === selectedLayout)?.pitboxes || track.pitboxes || '?'}</strong>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid rgba(255,255,255,0.05)', paddingBottom: '10px' }}>
                <span style={{ color: 'var(--text-muted)' }}>Layouts</span>
                <strong>{track.layouts?.length || 1}</strong>
              </div>
            </div>
            
            {track.description && (
              <div style={{ marginTop: '40px' }}>
                <h2>Description</h2>
                <p style={{ color: 'var(--text-muted)', lineHeight: '1.6', fontSize: '0.95rem' }}>{track.description}</p>
              </div>
            )}
          </div>

          {/* fast_lane.ai upload section */}
          <div style={{ backgroundColor: 'var(--sidebar-bg)', padding: '30px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
            <h2 style={{ marginTop: 0, borderBottom: '2px solid var(--primary-color)', paddingBottom: '10px' }}>AI Traffic (fast_lane.ai)</h2>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.95rem', marginBottom: '20px' }}>
              Add or update the AI trajectory so traffic can run on this track.
            </p>
            
            <div style={{ display: 'flex', flexDirection: 'column', gap: '15px' }}>
              <input 
                type="file" 
                accept=".ai,.aip" 
                ref={fileInputRef}
                style={{
                  padding: '10px',
                  backgroundColor: 'rgba(255,255,255,0.05)',
                  border: '1px solid var(--border-color)',
                  borderRadius: '6px',
                  color: 'white'
                }}
              />
              
              <button 
                onClick={async () => {
                  const file = fileInputRef.current?.files?.[0];
                  if (!file) {
                    setUploadMsg({ text: 'Please select an .ai or .aip file first', type: 'error' });
                    return;
                  }

                  setUploading(true);
                  setUploadMsg({ text: '', type: '' });

                  const formData = new FormData();
                  formData.append('file', file);
                  formData.append('trackId', trackId);
                  formData.append('layoutId', selectedLayout);

                  try {
                    const res = await fetch('/api/content/tracks/upload-fastlane', {
                      method: 'POST',
                      body: formData,
                    });
                    const data = await res.json();
                    
                    if (res.ok) {
                      setUploadMsg({ text: data.message, type: 'success' });
                      if (fileInputRef.current) fileInputRef.current.value = '';
                    } else {
                      setUploadMsg({ text: data.message, type: 'error' });
                    }
                  } catch (err: any) {
                    setUploadMsg({ text: 'Network error : ' + err.message, type: 'error' });
                  } finally {
                    setUploading(false);
                  }
                }}
                disabled={uploading}
                style={{ 
                  padding: '12px', 
                  backgroundColor: 'var(--primary-color)', 
                  color: 'white', 
                  border: 'none', 
                  borderRadius: '6px', 
                  fontWeight: 'bold', 
                  cursor: uploading ? 'not-allowed' : 'pointer',
                  opacity: uploading ? 0.7 : 1
                }}
              >
                {uploading ? 'Uploading...' : 'Upload fast_lane.ai'}
              </button>
              
              {uploadMsg.text && (
                <div style={{ 
                  padding: '10px', 
                  borderRadius: '6px', 
                  backgroundColor: uploadMsg.type === 'success' ? 'rgba(46, 204, 113, 0.2)' : 'rgba(231, 76, 60, 0.2)',
                  color: uploadMsg.type === 'success' ? '#2ecc71' : '#e74c3c',
                  border: `1px solid ${uploadMsg.type === 'success' ? '#2ecc71' : '#e74c3c'}`
                }}>
                  {uploadMsg.text}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function App() {
  const urlParams = new URLSearchParams(window.location.search);
  const viewCarId = urlParams.get('car');
  const viewTrackId = urlParams.get('track');
  const viewLive = urlParams.get('live') === 'true';

  if (viewCarId) return <StandaloneCarView carId={viewCarId} />;
  if (viewTrackId) return <StandaloneTrackView trackId={viewTrackId} />;
  if (viewLive) return <LiveTiming />;

  const [activeTab, setActiveTab] = useState<'status' | 'config' | 'content' | 'events' | 'users'>('status')
  const [apiStatus, setApiStatus] = useState<string>("Waiting...")
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false)
  const [username, setUsername] = useState<string>("")
  const [password, setPassword] = useState<string>("")
  const [newPassword, setNewPassword] = useState<string>("")
  const [requirePasswordChange, setRequirePasswordChange] = useState<boolean>(false)
  const [userRole, setUserRole] = useState<string>("")
  const [error, setError] = useState<string>("")
  const [loading, setLoading] = useState<boolean>(true)
  
  const [serverStatus, setServerStatus] = useState<string>("Unknown")
  const [isStatusLoading, setIsStatusLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState(false)
  const [consoleLogs, setConsoleLogs] = useState<string[]>([])
  const consoleEndRef = useRef<HTMLDivElement>(null)
  
  const [liveMetrics, setLiveMetrics] = useState<LiveMetrics | null>(null)

  useEffect(() => {
    fetch('/api/auth/verify')
      .then(res => res.json().then(data => ({ status: res.status, ok: res.ok, data })))
      .then(({ ok, data }) => {
        if (ok) {
          setIsAuthenticated(true)
          setUsername(data.username)
          setUserRole(data.role)
        } else {
          setIsAuthenticated(false)
        }
      })
      .catch(() => setIsAuthenticated(false))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (isAuthenticated) {
      fetch('/api/health')
        .then(res => res.json())
        .then(data => setApiStatus(data.message))
        .catch(err => setApiStatus("Backend connection error: " + err.message))
    }
  }, [isAuthenticated])

  useEffect(() => {
    if (!isAuthenticated || activeTab !== 'status') return

    const fetchStatus = () => {
      fetch('/api/server/status')
        .then(res => res.json())
        .then(data => {
          setServerStatus(data.status)
          setIsStatusLoading(false)
        })
        .catch(() => setServerStatus("Error"))
    }

    fetchStatus()
    const interval = setInterval(fetchStatus, 5000)
    return () => clearInterval(interval)
  }, [isAuthenticated, activeTab])

  // Live Metrics polling
  useEffect(() => {
    if (!isAuthenticated || activeTab !== 'status' || serverStatus !== 'Running') return

    const fetchMetrics = async () => {
      try {
        const res = await fetch('/api/server/live-metrics')
        if (res.ok) {
          const data = await res.json()
          setLiveMetrics(data)
        }
      } catch (err) {
        console.error(err)
      }
    }

    fetchMetrics()
    const metricsInterval = setInterval(fetchMetrics, 3000)
    return () => clearInterval(metricsInterval)
  }, [isAuthenticated, activeTab, serverStatus])

  // WebSocket for console logs
  useEffect(() => {
    if (!isAuthenticated || activeTab !== 'status') return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/server/console`
    
    let ws: WebSocket | null = null
    try {
      ws = new WebSocket(wsUrl)
      
      ws.onmessage = (event) => {
        setConsoleLogs(prev => {
          const newLogs = [...prev, event.data]
          if (newLogs.length > 500) return newLogs.slice(newLogs.length - 500)
          return newLogs
        })
      }
      ws.onerror = (error) => {
        console.error("WebSocket error:", error)
      }
    } catch (err) {
      console.error("Failed to connect WebSocket", err)
    }

    return () => {
      if (ws) ws.close()
    }
  }, [isAuthenticated, activeTab])

  useEffect(() => {
    if (consoleEndRef.current && activeTab === 'status') {
      consoleEndRef.current.scrollIntoView({ behavior: "smooth" })
    }
  }, [consoleLogs, activeTab])

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      })
      const data = await response.json()
      if (response.ok) {
        setIsAuthenticated(true)
      } else {
        if (response.status === 403 && data.require_password_change) {
          setRequirePasswordChange(true)
        } else {
          setError(data.message || "Password incorrect")
        }
      }
    } catch (err: any) {
      setError("Connection error: " + err.message)
    }
  }

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    try {
      const response = await fetch('/api/change-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, old_password: password, new_password: newPassword })
      })
      const data = await response.json()
      if (response.ok) {
        setRequirePasswordChange(false)
        setPassword(newPassword)
        const loginRes = await fetch('/api/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password: newPassword })
        })
        if (loginRes.ok) {
          setIsAuthenticated(true)
        }
      } else {
        setError(data.message || "Error")
      }
    } catch (err: any) {
      setError("Error : " + err.message)
    }
  }

  const handleServerAction = async (action: 'start' | 'stop') => {
    if (action === 'start') {
      setConsoleLogs([]); // Clear logs on start
    }
    setActionLoading(true)
    try {
      await fetch(`/api/server/${action}`, { method: 'POST' })
      const res = await fetch('/api/server/status')
      if (res.ok) {
        const data = await res.json()
        setServerStatus(data.status)
      }
    } catch (err) {
      console.error(err)
    } finally {
      setActionLoading(false)
    }
  }

  const handleLogout = async () => {
    try {
      await fetch('/api/logout', { method: 'POST' })
    } catch (e) {
      console.error(e)
    }
    setIsAuthenticated(false)
    setUsername("")
    setUserRole("")
    window.location.reload()
  }

  const handleRestart = async () => {
    setConsoleLogs([]); // Clear logs on restart
    setActionLoading(true)
    try {
      await fetch('/api/server/stop', { method: 'POST' })
      await new Promise(r => setTimeout(r, 1000))
      await fetch('/api/server/start', { method: 'POST' })
      const res = await fetch('/api/server/status')
      if (res.ok) {
        const data = await res.json()
        setServerStatus(data.status)
      }
    } catch (err) {
      console.error(err)
    } finally {
      setActionLoading(false)
    }
  }

  const getStatusBadge = () => {
    if (isStatusLoading) {
      return <div className="skeleton" style={{ width: '100px', height: '28px', borderRadius: '16px' }}></div>
    }

    if (serverStatus === "Running") {
      return (
        <div className="status-badge status-online">
          <div className="status-dot"></div> Online
        </div>
      )
    } else if (serverStatus === "Stopped") {
      return (
        <div className="status-badge status-offline">
          <div className="status-dot"></div> Offline
        </div>
      )
    } else {
      return (
        <div className="status-badge status-unknown">
          <div className="status-dot"></div> Unknown
        </div>
      )
    }
  }

  const formatTime = (ms: number) => {
    if (!ms) return "-"
    const min = Math.floor(ms / 60000)
    const sec = Math.floor((ms % 60000) / 1000)
    const msRemainder = ms % 1000
    return `${min}:${sec.toString().padStart(2, '0')}.${msRemainder.toString().padStart(3, '0')}`
  }

  if (loading) {
    return <div className="login-container">Loading...</div>
  }

  if (!isAuthenticated) {
    if (requirePasswordChange) {
      return (
        <div className="login-container">
          <div className="login-box">
            <h2>New password</h2>
            <p style={{ color: 'var(--text-muted)', marginBottom: '20px', fontSize: '0.9rem' }}>
              For security reasons, please set your personal password.
            </p>
            <form onSubmit={handleChangePassword}>
              <div className="input-group">
                <input 
                  type="password" 
                  placeholder="New password" 
                  value={newPassword}
                  onChange={e => setNewPassword(e.target.value)}
                  autoFocus
                />
              </div>
              {error && <p className="error-msg">{error}</p>}
              <button type="submit" className="login-btn">Save</button>
            </form>
          </div>
        </div>
      )
    }

    return (
      <div className="login-container">
        <div className="login-box">
          <h2>Web GUI ACServer</h2>
          <form onSubmit={handleLogin}>
            <div className="input-group" style={{ marginBottom: '15px' }}>
              <input 
                type="text" 
                placeholder="Username" 
                value={username}
                onChange={e => setUsername(e.target.value)}
                autoFocus
              />
            </div>
            <div className="input-group">
              <input 
                type="password" 
                placeholder="Password" 
                value={password}
                onChange={e => setPassword(e.target.value)}
              />
            </div>
            {error && <p className="error-msg">{error}</p>}
            <button type="submit" className="login-btn">Login</button>
          </form>
        </div>
      </div>
    )
  }

  return (
    <div className="app-container">
      <aside className="sidebar">
        <div className="logo">Web GUI ACServer</div>
        <nav className="nav-menu">
          <a 
            href="#" 
            className={`nav-item ${activeTab === 'status' ? 'active' : ''}`}
            onClick={(e) => { e.preventDefault(); setActiveTab('status') }}
          >
            Dashboard
          </a>
          <a 
            href="#" 
            className={`nav-item ${activeTab === 'events' ? 'active' : ''}`}
            onClick={(e) => { e.preventDefault(); setActiveTab('events') }}
          >
            Events
          </a>
          <a 
            href="#" 
            className={`nav-item ${activeTab === 'config' ? 'active' : ''}`}
            onClick={(e) => { e.preventDefault(); setActiveTab('config') }}
          >
            Configuration
          </a>
          <div style={{ marginLeft: '16px', display: 'flex', flexDirection: 'column' }}>
            {userRole === 'admin' && (
              <a 
                href="#" 
                className={`nav-item ${activeTab === 'users' ? 'active' : ''}`}
                style={{ fontSize: '0.9rem', padding: '8px 16px', opacity: 0.8 }}
                onClick={(e) => { e.preventDefault(); setActiveTab('users') }}
              >
                ↳ Users
              </a>
            )}
            <a 
              href="#" 
              className={`nav-item ${activeTab === 'plugins' ? 'active' : ''}`}
              style={{ fontSize: '0.9rem', padding: '8px 16px', opacity: 0.8 }}
              onClick={(e) => { e.preventDefault(); setActiveTab('plugins') }}
            >
              ↳ Plugins
            </a>
            <a 
              href="#" 
              className={`nav-item ${activeTab === 'logs' ? 'active' : ''}`}
              style={{ fontSize: '0.9rem', padding: '8px 16px', opacity: 0.8 }}
              onClick={(e) => { e.preventDefault(); setActiveTab('logs') }}
            >
              ↳ Global Logs
            </a>
          </div>
          <a 
            href="#" 
            className={`nav-item ${activeTab === 'content' ? 'active' : ''}`}
            onClick={(e) => { e.preventDefault(); setActiveTab('content') }}
          >
            Content
          </a>
          <a 
            href="?live=true" 
            target="_blank"
            className="nav-item"
            style={{ marginTop: '20px', color: '#f39c12' }}
          >
            Live Timing ↗
          </a>
        </nav>
      </aside>
      <main className="main-content">
        <header className="top-header">
          <h1>
            {activeTab === 'status' && 'Dashboard'}
            {activeTab === 'events' && 'Event Builder'}
            {activeTab === 'config' && 'Configuration'}
            {activeTab === 'users' && 'User Management'}
            {activeTab === 'plugins' && 'Plugin Management'}
            {activeTab === 'logs' && 'Global Logs'}
            {activeTab === 'content' && 'Content Management'}
          </h1>
          <div style={{ display: 'flex', alignItems: 'center', gap: '15px' }}>
            <div style={{ color: 'var(--text-muted)' }}>
              Logged in as: <strong style={{ color: 'var(--primary-color)' }}>{username}</strong>
            </div>
            <button 
              className="logout-btn" 
              onClick={handleLogout}
            >
              Logout
            </button>
          </div>
        </header>
        
        <div className="content-area">
          {activeTab === 'status' && (
            <div className="card" style={{ maxWidth: '800px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                <h2 style={{ margin: 0 }}>Server Status</h2>
                {getStatusBadge()}
              </div>
              
              <div className="control-buttons">
                {(serverStatus === "Stopped" || serverStatus === "Crashed") && (
                  <button 
                    className="btn-start" 
                    onClick={() => handleServerAction('start')}
                    disabled={actionLoading}
                  >
                    {actionLoading ? "Loading..." : "Start"}
                  </button>
                )}
                {serverStatus === "Running" && (
                  <>
                    <button 
                      className="btn-stop" 
                      onClick={() => handleServerAction('stop')}
                      disabled={actionLoading}
                    >
                      {actionLoading ? "Loading..." : "Stop"}
                    </button>
                    <button 
                      className="btn-restart" 
                      onClick={handleRestart}
                      disabled={actionLoading}
                    >
                      {actionLoading ? "Loading..." : "Restart"}
                    </button>
                  </>
                )}
              </div>

              <div className="console-widget">
                {consoleLogs.map((log, i) => (
                  <p key={i} className="console-line">{log}</p>
                ))}
                <div ref={consoleEndRef} />
              </div>

              <div className="card" style={{ marginTop: '20px', backgroundColor: '#1a1d24', border: '1px solid #2d313a' }}>
                <h3 style={{ marginTop: 0 }}>Live Metrics</h3>
                {serverStatus === 'Running' && liveMetrics ? (
                  <div>
                    <div style={{ display: 'flex', gap: '20px', marginBottom: '20px' }}>
                      <div style={{ background: '#22262e', padding: '16px', borderRadius: '8px', flex: 1 }}>
                        <div style={{ fontSize: '0.9rem', color: '#8b949e', marginBottom: '4px' }}>Connected Drivers</div>
                        <div style={{ fontSize: '1.5rem', fontWeight: 600 }}>{liveMetrics.info.clients} / {liveMetrics.info.maxclients}</div>
                      </div>
                      <div style={{ background: '#22262e', padding: '16px', borderRadius: '8px', flex: 1 }}>
                        <div style={{ fontSize: '0.9rem', color: '#8b949e', marginBottom: '4px' }}>Quick Access</div>
                        <div>
                          <a href={`https://acstuff.ru/s/q:race/online/join?ip=${liveMetrics.info.ip}&httpPort=${liveMetrics.info.cport}`}
                             target="_blank"
                             rel="noopener noreferrer"
                             className="btn-save" 
                             style={{ padding: '6px 12px', textDecoration: 'none', display: 'inline-block', marginTop: '4px' }}>
                            Join (Content Manager)
                          </a>
                        </div>
                      </div>
                    </div>
                    
                    <h4 style={{ marginBottom: '12px', borderBottom: '1px solid #2d313a', paddingBottom: '8px' }}>Leaderboard / Track</h4>
                    <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                      <thead>
                        <tr style={{ color: '#8b949e', borderBottom: '1px solid #2d313a' }}>
                          <th style={{ padding: '8px 4px' }}>Driver</th>
                          <th style={{ padding: '8px 4px' }}>Car</th>
                          <th style={{ padding: '8px 4px' }}>Best Lap</th>
                        </tr>
                      </thead>
                      <tbody>
                        {liveMetrics.cars.filter(c => c.isConnected).length > 0 ? (
                          liveMetrics.cars.filter(c => c.isConnected)
                           .sort((a, b) => (a.bestLap || 99999999) - (b.bestLap || 99999999))
                           .map((car, idx) => (
                            <tr key={idx} style={{ borderBottom: '1px solid #2d313a' }}>
                              <td style={{ padding: '10px 4px' }}>{car.driverName || 'Unknown'} {car.driverTeam ? `(${car.driverTeam})` : ''}</td>
                              <td style={{ padding: '10px 4px', fontSize: '0.9rem' }}>{car.model}</td>
                              <td style={{ padding: '10px 4px', fontFamily: 'monospace', fontSize: '1rem' }}>{formatTime(car.bestLap || 0)}</td>
                            </tr>
                          ))
                        ) : (
                          <tr><td colSpan={3} style={{textAlign: 'center', opacity: 0.6, padding: '20px'}}>No driver on track</td></tr>
                        )}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <p style={{ opacity: 0.6, margin: 0 }}>The server must be online to display metrics.</p>
                )}
              </div>
            </div>
          )}
        
        {activeTab === 'events' && (
          <EventBuilderView onLaunch={() => {
            setConsoleLogs([]);
            setActiveTab('status');
          }} />
        )}

        {activeTab === 'config' && (
          <>
            <GeneralConfigView />
            {userRole === 'admin' && <PublicModeConfigView />}
            <SystemUpgradeView />
          </>
        )}

        {activeTab === 'users' && <UsersView />}
        
        {activeTab === 'plugins' && (
          <PluginsView />
        )}
        
        {activeTab === 'logs' && (
          <LogsView />
        )}
        
        {activeTab === 'content' && <ContentConfigView />}

        </div>
      </main>
    </div>
  )
}

export default App
