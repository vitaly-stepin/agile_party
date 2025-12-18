import { Card, Button } from '../common';
import { useRoom } from '../../context/RoomContext';

interface ResultsDisplayProps {
  onClear: () => void;
}

export default function ResultsDisplay({ onClear }: ResultsDisplayProps) {
  const { roomState } = useRoom();

  const votes = roomState?.votes || [];
  const average = roomState?.average;

  // Group votes by value for better visualization
  const voteCounts = votes.reduce((acc, vote) => {
    const value = vote.value;
    acc[value] = (acc[value] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const sortedVoteValues = Object.keys(voteCounts).sort((a, b) => {
    // Put '?' at the end
    if (a === '?') return 1;
    if (b === '?') return -1;
    return parseFloat(a) - parseFloat(b);
  });

  return (
    <div className="space-y-6">
      {/* Average Display */}
      <Card variant="elevated" padding="lg">
        <div className="text-center">
          <h3 className="text-sm font-medium text-gray-600 mb-2">
            Average Estimate
          </h3>
          <div className="text-6xl font-bold text-blue-600 mb-2">
            {average !== null && average !== undefined
              ? average.toFixed(1)
              : 'N/A'}
          </div>
          {average === null && (
            <p className="text-sm text-gray-500">
              No numeric votes to calculate average
            </p>
          )}
        </div>
      </Card>

      {/* Individual Votes */}
      <Card variant="outlined" padding="lg">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">
          All Votes
        </h3>

        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4 mb-6">
          {votes.map((vote, index) => (
            <div
              key={index}
              className="bg-gray-50 rounded-lg p-4 border border-gray-200"
            >
              <div className="text-2xl font-bold text-gray-900 text-center mb-1">
                {vote.value}
              </div>
              <div className="text-sm text-gray-600 text-center truncate">
                {vote.userName}
              </div>
            </div>
          ))}
        </div>

        {/* Vote Distribution */}
        <div className="border-t border-gray-200 pt-4">
          <h4 className="text-sm font-medium text-gray-700 mb-3">
            Vote Distribution
          </h4>
          <div className="space-y-2">
            {sortedVoteValues.map((value) => (
              <div key={value} className="flex items-center gap-3">
                <div className="w-12 text-right font-semibold text-gray-900">
                  {value}
                </div>
                <div className="flex-1 bg-gray-200 rounded-full h-6 overflow-hidden">
                  <div
                    className="bg-blue-500 h-full flex items-center justify-end px-2 text-white text-xs font-medium transition-all"
                    style={{
                      width: `${(voteCounts[value] / votes.length) * 100}%`,
                    }}
                  >
                    {voteCounts[value] > 1 && voteCounts[value]}
                  </div>
                </div>
                <div className="w-16 text-sm text-gray-600">
                  {voteCounts[value]} vote{voteCounts[value] !== 1 ? 's' : ''}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Clear Button */}
        <div className="mt-6 pt-4 border-t border-gray-200">
          <Button onClick={onClear} variant="primary" fullWidth>
            Clear Votes & Start New Round
          </Button>
        </div>
      </Card>
    </div>
  );
}
