import { useState } from 'react';
import { Card, Button } from '../common';
import { useRoom } from '../../context/RoomContext';
import { VALID_VOTES } from '../../types';
import type { VoteValue, ClientEvent } from '../../types';
import VoteCard from './VoteCard';

interface VotePanelProps {
  onReveal: () => void;
  sendEvent: (event: ClientEvent) => void;
}

export default function VotePanel({ onReveal, sendEvent }: VotePanelProps) {
  const { currentUser, roomState } = useRoom();
  const [selectedVote, setSelectedVote] = useState<VoteValue | null>(null);

  const handleVoteClick = (value: VoteValue) => {
    setSelectedVote(value);
    sendEvent({
      type: 'vote',
      payload: { value },
    });
  };

  const currentUserVoted = currentUser
    ? roomState?.users.some(
        (u) => u.id === currentUser.id && u.isVoted
      ) || false
    : false;

  const allUsersVoted = roomState?.users.every((u) => u.isVoted) || false;
  const hasUsers = (roomState?.users.length || 0) > 0;

  return (
    <div className="space-y-6">
      {/* Vote Cards Grid */}
      <Card variant="outlined" padding="lg">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">
          Select Your Estimate
        </h3>

        <div className="grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 gap-3 mb-6">
          {VALID_VOTES.map((vote) => (
            <VoteCard
              key={vote}
              value={vote}
              isSelected={selectedVote === vote}
              onClick={handleVoteClick}
              disabled={!currentUser}
            />
          ))}
        </div>

        {/* Vote Status */}
        <div className="flex items-center justify-between pt-4 border-t border-gray-200">
          <div className="text-sm text-gray-600">
            {currentUserVoted ? (
              <span className="text-green-600 font-medium">
                ✓ You voted {selectedVote}
              </span>
            ) : (
              <span>Select a card to vote</span>
            )}
          </div>

          <Button
            onClick={onReveal}
            disabled={!hasUsers || !allUsersVoted}
            variant="primary"
          >
            Reveal Votes
          </Button>
        </div>

        {!allUsersVoted && hasUsers && (
          <p className="text-xs text-gray-500 mt-2 text-right">
            Waiting for all users to vote...
          </p>
        )}
      </Card>

      {/* Info Card */}
      <Card variant="default" padding="md">
        <div className="text-sm text-gray-600">
          <p className="mb-2">
            <strong>How it works:</strong>
          </p>
          <ul className="list-disc list-inside space-y-1 text-xs">
            <li>Click a card to submit your vote</li>
            <li>You can change your vote before reveal</li>
            <li>Votes are hidden until everyone has voted</li>
            <li>Click "Reveal Votes" to see all estimates and the average</li>
          </ul>
        </div>
      </Card>
    </div>
  );
}
