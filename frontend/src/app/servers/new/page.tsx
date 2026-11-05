'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import axios from 'axios'

export default function CreateServerPage() {
  const router = useRouter()
  const [step, setStep] = useState(1)
  const [loading, setLoading] = useState(false)
  
  const [formData, setFormData] = useState({
    name: '',
    jar_path: 'server.jar',
    java_version: 'java17',
    xms_mb: 2048,
    xmx_mb: 4096,
    port: 25565,
    eula_accepted: false,
  })

  const handleNext = () => setStep((prev) => Math.min(prev + 1, 4))
  const handleBack = () => setStep((prev) => Math.max(prev - 1, 1))

  const handleSubmit = async () => {
    setLoading(true)
    try {
      await axios.post('/api/v1/servers', formData)
      router.push('/')
    } catch (err) {
      console.error('Failed to create server:', err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="container mx-auto max-w-2xl">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-3xl font-bold">Create New Server</h1>
          <button
            onClick={() => router.push('/')}
            className="px-4 py-2 text-gray-400 hover:text-white"
          >
            Cancel
          </button>
        </div>

        {/* Progress */}
        <div className="mb-8 flex justify-center space-x-2">
          {[1, 2, 3, 4].map((s) => (
            <div
              key={s}
              className={`h-2 w-12 rounded-full ${
                s <= step ? 'bg-blue-500' : 'bg-gray-700'
              }`}
            />
          ))}
        </div>

        <div className="bg-gray-800 rounded-lg p-8">
          {step === 1 && (
            <div className="space-y-4">
              <h2 className="text-xl font-semibold mb-4">Server Settings</h2>
              <div>
                <label className="block text-gray-300 mb-2">Server Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="my-server"
                  className="w-full px-4 py-2 bg-gray-700 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-gray-300 mb-2">Jar Path</label>
                <input
                  type="text"
                  value={formData.jar_path}
                  onChange={(e) => setFormData({ ...formData, jar_path: e.target.value })}
                  className="w-full px-4 py-2 bg-gray-700 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4">
              <h2 className="text-xl font-semibold mb-4">Java Configuration</h2>
              <div>
                <label className="block text-gray-300 mb-2">Java Version</label>
                <select
                  value={formData.java_version}
                  onChange={(e) => setFormData({ ...formData, java_version: e.target.value })}
                  className="w-full px-4 py-2 bg-gray-700 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="java8">Java 8</option>
                  <option value="java11">Java 11</option>
                  <option value="java17">Java 17</option>
                  <option value="java21">Java 21</option>
                </select>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-4">
              <h2 className="text-xl font-semibold mb-4">Resource Allocation</h2>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-gray-300 mb-2">Xms (Min MB)</label>
                  <input
                    type="number"
                    value={formData.xms_mb}
                    onChange={(e) => setFormData({ ...formData, xms_mb: parseInt(e.target.value) })}
                    className="w-full px-4 py-2 bg-gray-700 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-gray-300 mb-2">Xmx (Max MB)</label>
                  <input
                    type="number"
                    value={formData.xmx_mb}
                    onChange={(e) => setFormData({ ...formData, xmx_mb: parseInt(e.target.value) })}
                    className="w-full px-4 py-2 bg-gray-700 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
              <div>
                <label className="block text-gray-300 mb-2">Port</label>
                <input
                  type="number"
                  value={formData.port}
                  onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) })}
                  className="w-full px-4 py-2 bg-gray-700 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
          )}

          {step === 4 && (
            <div className="space-y-4">
              <h2 className="text-xl font-semibold mb-4">Review & EULA</h2>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-gray-300">Name:</span>
                  <span className="font-medium">{formData.name}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-300">Jar:</span>
                  <span className="font-medium">{formData.jar_path}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-300">Java:</span>
                  <span className="font-medium">{formData.java_version}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-300">Memory:</span>
                  <span className="font-medium">{formData.xms_mb}MB / {formData.xmx_mb}MB</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-300">Port:</span>
                  <span className="font-medium">{formData.port}</span>
                </div>
              </div>
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  checked={formData.eula_accepted}
                  onChange={(e) => setFormData({ ...formData, eula_accepted: e.target.checked })}
                  className="w-4 h-4"
                />
                <label className="text-gray-300">I accept the Minecraft EULA</label>
              </div>
            </div>
          )}

          <div className="mt-8 flex justify-between">
            {step > 1 && (
              <button
                onClick={handleBack}
                className="px-6 py-2 bg-gray-700 hover:bg-gray-600 rounded transition"
              >
                Back
              </button>
            )}
            {step < 4 ? (
              <button
                onClick={handleNext}
                disabled={!formData.name}
                className="px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded disabled:opacity-50 transition"
              >
                Next
              </button>
            ) : (
              <button
                onClick={handleSubmit}
                disabled={loading || !formData.eula_accepted}
                className="px-6 py-2 bg-green-600 hover:bg-green-700 rounded disabled:opacity-50 transition"
              >
                {loading ? 'Creating...' : 'Create Server'}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
