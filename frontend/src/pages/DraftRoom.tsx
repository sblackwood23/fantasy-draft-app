import { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useWebSocket } from '../hooks/useWebSocket';
import { useDraftStore } from '../store/draftStore';
import { usePlayerStore } from '../store/playerStore';
import { PlayerList } from '../components/PlayerList';

export function DraftRoom() {
  // URL query params
  const [searchParams] = useSearchParams();
  const userId = Number(searchParams.get('userId'));

  // Custom hook - WebSocket connection methods
  const { connect, disconnect } = useWebSocket();

  // Zustand store selectors - each subscribes to a slice of global state
  const connectionStatus = useDraftStore((s) => s.connectionStatus);
  const eventID = useDraftStore((s) => s.eventID);
  const draftStatus = useDraftStore((s) => s.draftStatus);
  const currentTurn = useDraftStore((s) => s.currentTurn);
  const roundNumber = useDraftStore((s) => s.roundNumber);
  const pickHistory = useDraftStore((s) => s.pickHistory);
  const lastError = useDraftStore((s) => s.lastError);
  const turnDeadline = useDraftStore((s) => s.turnDeadline);

  const initializeEventPlayers = usePlayerStore((s) => s.setEventPlayers);

  // Effect: connect WebSocket on mount, disconnect on unmount
  useEffect(() => {
    connect();
    return () => disconnect(); // cleanup function
  }, [connect, disconnect]);

  // Initialize players for the given eventID
  useEffect(() => {
    if (eventID !== null) {
      initializeEventPlayers(eventID)
    }
  }, [eventID, initializeEventPlayers])

  // Computed value - derived from state, recalculates each render
  const isMyTurn = currentTurn === userId;

  return (
    <div className="min-h-screen bg-gray-900 text-white p-4">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold">Draft Room</h1>
        </div>

        {/* Connection Status */}
        <div className="mb-4 p-3 bg-gray-800 rounded">
          <div className="flex items-center gap-2">
            <span
              className={`w-3 h-3 rounded-full ${
                connectionStatus === 'connected'
                  ? 'bg-green-500'
                  : connectionStatus === 'connecting'
                  ? 'bg-yellow-500'
                  : 'bg-red-500'
              }`}
            />
            <span className="text-sm">
              {connectionStatus === 'connected'
                ? 'Connected'
                : connectionStatus === 'connecting'
                ? 'Connecting...'
                : 'Disconnected'}
            </span>
          </div>
          <p className="text-xs text-gray-400 mt-1">Your User ID: {userId}</p>
        </div>

        {/* Error Display */}
        {lastError && (
          <div className="mb-4 p-3 bg-red-900/50 border border-red-700 rounded text-red-300">
            Error: {lastError}
          </div>
        )}

        {/* Draft Status */}
        <div className="mb-4 p-4 bg-gray-800 rounded">
          <h2 className="font-semibold mb-2">Draft Status</h2>
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div>Status: <span className="text-blue-400">{draftStatus}</span></div>
            <div>Round: <span className="text-blue-400">{roundNumber}</span></div>
            <div>Current Turn: <span className="text-blue-400">{currentTurn ?? 'N/A'}</span></div>
            <div>
              {isMyTurn ? (
                <span className="text-green-400 font-bold">YOUR TURN!</span>
              ) : (
                <span className="text-gray-400">Waiting...</span>
              )}
            </div>
          </div>
          {turnDeadline && (
            <div className="mt-2 text-xs text-gray-400">
              Turn deadline: {new Date(turnDeadline * 1000).toLocaleTimeString()}
            </div>
          )}
        </div>

        {/* Player List */}
        <div className="mb-4">
          <PlayerList />
        </div>

        {/* Pick History */}
        <div className="p-4 bg-gray-800 rounded">
          <h2 className="font-semibold mb-2">
            Pick History ({pickHistory.length})
          </h2>
          {pickHistory.length === 0 ? (
            <p className="text-gray-400 text-sm">No picks yet.</p>
          ) : (
            <div className="space-y-1 text-sm max-h-48 overflow-y-auto">
              {pickHistory.map((pick, i) => (
                <div key={i} className="flex justify-between">
                  <span>
                    #{pick.pickNumber} - User {pick.userID} picked Player {pick.playerID}
                  </span>
                  <span className="text-gray-400">
                    Round {pick.round}
                    {pick.autoDraft && (
                      <span className="text-yellow-500 ml-2">(auto)</span>
                    )}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Debug: Raw State */}
        <details className="mt-4">
          <summary className="text-gray-400 text-sm cursor-pointer">
            Debug: Raw Store State
          </summary>
          <pre className="mt-2 p-4 bg-gray-800 rounded text-xs overflow-x-auto">
            {JSON.stringify(useDraftStore.getState(), null, 2)}
          </pre>
        </details>
      </div>
    </div>
  );
}
