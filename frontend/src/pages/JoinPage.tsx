import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { joinDraft } from '../api/client';

export function JoinPage() {
  const navigate = useNavigate();
  const [teamName, setTeamName] = useState<string>('');
  const [passKey, setPassKey] = useState<string>('');
  const [error, setError] = useState<string | null>(null);

  function handleJoin() {
    if (!teamName || !passKey) return;
    // Clear out error before attempting to join draft
    setError(null);
    joinDraft(teamName, passKey)
      .then(() => navigate('/draft'))
      .catch((err: Error) => setError(err.message || 'Failed to join draft'));
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-900 text-white">
      <form
        onSubmit={(e) => { e.preventDefault(); handleJoin(); }}
        className="bg-gray-800 p-8 rounded-lg shadow-lg max-w-md w-full"
      >
        <h1 className="text-2xl font-bold mb-6 text-center">Join Draft</h1>

        <div className="mb-6">
          <label className="block text-lg font-medium mb-2">Team Name</label>
          <input
            type="text"
            value={teamName}
            onChange={(e) => setTeamName(e.target.value)}
            placeholder="Enter Team Name"
            className="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
          />
        </div>

        <div className="mb-6">
          <label className="block text-lg font-medium mb-2">Passkey</label>
          <input
            type="text"
            value={passKey}
            onChange={(e) => setPassKey(e.target.value)}
            placeholder="Enter Passkey"
            className="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
          />
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-900/50 border border-red-500 rounded-lg text-red-300 text-sm">
            {error}
          </div>
        )}

        <button
          type="submit"
          disabled={!(teamName && passKey)}
          className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:text-gray-400 disabled:cursor-not-allowed text-white font-medium py-2 px-4 rounded transition-colors"
        >
          Join
        </button>
      </form>
    </div>
  );
}
