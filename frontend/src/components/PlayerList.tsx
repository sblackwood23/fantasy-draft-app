import { useState } from "react";
import { usePlayerStore } from "../store/playerStore";
import { useDraftStore } from "../store/draftStore";
import type { Player } from "../types";

export function PlayerList() {
  const [searchFilter, setSearchFilter] = useState<string>('');
  const [countryCodeFilter, setCountryCodeFilter] = useState<string[] | null>(null);
  const [includeDrafted, setIncludeDrafted] = useState<boolean>(false);

  const eventPlayers = usePlayerStore((s) => s.eventPlayers);
  const availablePlayerIDs = useDraftStore((s) => s.availablePlayerIDs);
  // Derive availablePlayers from availablePlayerIDs - reconcile against eventPlayers
  const availablePlayers = eventPlayers.filter((p) => availablePlayerIDs.some((id) => id === p.id));
  const draftedPlayers = eventPlayers.filter((p) => !availablePlayerIDs.some((id) => id === p.id))

  const displayedPlayers = getDisplayedPlayers();

  function getDisplayedPlayers(): Player[] {
    let players = includeDrafted ? eventPlayers : availablePlayers;

    if (searchFilter) {
      const query = searchFilter.toLowerCase();
      players = players.filter((p) =>
        `${p.first_name} ${p.last_name}`.toLowerCase().includes(query)
      );
    }

    if (countryCodeFilter && countryCodeFilter.length > 0) {
      players = players.filter((p) => countryCodeFilter.includes(p.countryCode));
    }

    return players;
  }

  // Derive unique country codes from all event players for the filter options
  const countryCodes = [...new Set(eventPlayers.map((p) => p.countryCode))].sort();

  function toggleCountryCode(code: string) {
    setCountryCodeFilter((prev) => {
      if (!prev) return [code];
      if (prev.includes(code)) {
        const next = prev.filter((c) => c !== code);
        return next.length === 0 ? null : next;
      }
      return [...prev, code];
    });
  }

  return (
    <div className="p-4 bg-gray-800 rounded">
      <div className="flex items-center justify-between mb-4">
        <h2 className="font-semibold">Players ({displayedPlayers.length})</h2>
        <label className="flex items-center gap-2 text-sm text-gray-400">
          <input
            type="checkbox"
            checked={includeDrafted}
            onChange={(e) => setIncludeDrafted(e.target.checked)}
            className="rounded"
          />
          Show drafted
        </label>
      </div>

      <input
        type="text"
        value={searchFilter}
        onChange={(e) => setSearchFilter(e.target.value)}
        placeholder="Search players..."
        className="w-full mb-3 px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm text-white placeholder-gray-400 focus:outline-none focus:border-blue-500"
      />

      {countryCodes.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-4">
          {countryCodes.map((code) => (
            <button
              key={code}
              onClick={() => toggleCountryCode(code)}
              className={`px-2 py-1 rounded text-xs ${
                countryCodeFilter?.includes(code)
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 text-gray-400 hover:bg-gray-600'
              }`}
            >
              {code}
            </button>
          ))}
          {countryCodeFilter && (
            <button
              onClick={() => setCountryCodeFilter(null)}
              className="px-2 py-1 rounded text-xs bg-gray-700 text-red-400 hover:bg-gray-600"
            >
              Clear
            </button>
          )}
        </div>
      )}

      {displayedPlayers.length === 0 ? (
        <p className="text-gray-400 text-sm">No players loaded.</p>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-gray-400 border-b border-gray-700">
              <th className="pb-2">Name</th>
              <th className="pb-2">Country</th>
              <th className="pb-2">Status</th>
            </tr>
          </thead>
          <tbody>
            {displayedPlayers.map((player) => {
              const isDrafted = draftedPlayers.some((d) => d.id === player.id);
              return (
                <tr
                  key={player.id}
                  className={`border-b border-gray-700/50 ${isDrafted ? 'opacity-40' : ''}`}
                >
                  <td className="py-2">
                    {player.first_name} {player.last_name}
                    {isDrafted && <span className="ml-2 text-xs text-red-400">(drafted)</span>}
                  </td>
                  <td className="py-2">{player.countryCode}</td>
                  <td className="py-2">{player.status}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
}