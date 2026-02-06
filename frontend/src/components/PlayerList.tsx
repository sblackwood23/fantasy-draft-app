import { useState, useRef, useEffect } from "react";
import { usePlayerStore } from "../store/playerStore";
import { useDraftStore } from "../store/draftStore";
import type { Player } from "../types";

export function PlayerList() {
  const [searchFilter, setSearchFilter] = useState<string>('');
  const [countryCodeFilter, setCountryCodeFilter] = useState<string[] | null>(null);
  const [countryDropdownOpen, setCountryDropdownOpen] = useState(false);
  const [playerFilter, setPlayerFilter] = useState<'available' | 'drafted' | 'all'>('available');
  const countryDropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (countryDropdownRef.current && !countryDropdownRef.current.contains(e.target as Node)) {
        setCountryDropdownOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const eventPlayers = usePlayerStore((s) => s.eventPlayers);
  const availablePlayerIDs = useDraftStore((s) => s.availablePlayerIDs);
  // null = no draft data yet, all players are available
  const availablePlayers = availablePlayerIDs === null
    ? eventPlayers
    : eventPlayers.filter((p) => availablePlayerIDs.some((id) => id === p.id));
  const draftedPlayers = availablePlayerIDs === null
    ? []
    : eventPlayers.filter((p) => !availablePlayerIDs.some((id) => id === p.id));

  const displayedPlayers = getDisplayedPlayers();

  function getDisplayedPlayers(): Player[] {
    let players = playerFilter === 'all' ? eventPlayers
      : playerFilter === 'drafted' ? draftedPlayers
      : availablePlayers;

    if (searchFilter) {
      const query = searchFilter.toLowerCase();
      players = players.filter((p) =>
        `${p.firstName} ${p.lastName}`.toLowerCase().includes(query)
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
    <div className="p-4 bg-gray-800 rounded flex flex-col max-h-[70vh]">
      <div className="flex items-center justify-between mb-4">
        <h2 className="font-semibold">Players ({displayedPlayers.length})</h2>
        <select
          value={playerFilter}
          onChange={(e) => setPlayerFilter(e.target.value as 'available' | 'drafted' | 'all')}
          className="px-2 py-1 bg-gray-700 border border-gray-600 rounded text-sm text-white"
        >
          <option value="available">Available</option>
          <option value="drafted">Drafted</option>
          <option value="all">All</option>
        </select>
      </div>

      <div className="flex gap-2 mb-4">
        <input
          type="text"
          value={searchFilter}
          onChange={(e) => setSearchFilter(e.target.value)}
          placeholder="Search players..."
          className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm text-white placeholder-gray-400 focus:outline-none focus:border-blue-500"
        />

      {countryCodes.length > 0 && (
        <div className="relative w-48" ref={countryDropdownRef}>
          <button
            onClick={() => setCountryDropdownOpen((prev) => !prev)}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm text-left flex items-center justify-between"
          >
            <span className={`truncate ${countryCodeFilter ? 'text-white' : 'text-gray-400'}`}>
              {countryCodeFilter
                ? countryCodeFilter.join(', ')
                : 'Filter by country'}
            </span>
            <span className="text-gray-400 text-xs">{countryDropdownOpen ? '\u25B2' : '\u25BC'}</span>
          </button>
          {countryDropdownOpen && (
            <div className="absolute z-10 mt-1 w-full bg-gray-700 border border-gray-600 rounded shadow-lg max-h-48 overflow-y-auto">
              {countryCodeFilter && (
                <div className="flex border-b border-gray-600">
                  <button
                    onClick={() => setCountryCodeFilter(null)}
                    className="flex-1 px-3 py-2 text-sm text-red-400 hover:bg-gray-600 text-left"
                  >
                    Clear
                  </button>
                  <button
                    onClick={() => {
                      const inverted = countryCodes.filter((c) => !countryCodeFilter.includes(c));
                      setCountryCodeFilter(inverted.length > 0 ? inverted : null);
                    }}
                    className="flex-1 px-3 py-2 text-sm text-blue-400 hover:bg-gray-600 text-left"
                  >
                    Invert
                  </button>
                </div>
              )}
              {countryCodes.map((code) => (
                <label
                  key={code}
                  className="flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-600 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={countryCodeFilter?.includes(code) ?? false}
                    onChange={() => toggleCountryCode(code)}
                    className="rounded"
                  />
                  {code}
                </label>
              ))}
            </div>
          )}
        </div>
      )}
      </div>

      {displayedPlayers.length === 0 ? (
        <p className="text-gray-400 text-sm">No players loaded.</p>
      ) : (
        <div className="overflow-y-auto flex-1 min-h-0">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-gray-800">
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
                      {player.firstName} {player.lastName}
                      {isDrafted && <span className="ml-2 text-xs text-red-400">(drafted)</span>}
                    </td>
                    <td className="py-2">{player.countryCode}</td>
                    <td className="py-2">{player.status}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}