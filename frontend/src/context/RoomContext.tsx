import React, { createContext, useContext, useState, useCallback, useMemo } from 'react';
import type { ReactNode } from 'react';
import type { Room, User, Vote, RoomState, CreateRoomRequest } from '../types';
import { api } from '../services/api';

interface RoomContextState {
  // Room data
  room: Room | null;
  roomState: RoomState | null;

  // Connection state
  isLoading: boolean;
  error: string | null;

  // Current user
  currentUserId: string | null;
  currentUser: User | null;

  // Actions
  createRoom: (roomName: string, nickname: string) => Promise<string>;
  joinRoom: (roomId: string, nickname: string) => Promise<void>;
  leaveRoom: () => Promise<void>;
  setRoomState: (state: RoomState) => void;
  updateUsers: (users: User[]) => void;
  updateUserVoteStatus: (userId: string, hasVoted: boolean) => void;
  updateVotes: (votes: Vote[]) => void;
  setRevealed: (revealed: boolean, average?: number | null) => void;
  clearError: () => void;
}

const RoomContext = createContext<RoomContextState | undefined>(undefined);

interface RoomProviderProps {
  children: ReactNode;
}

export const RoomProvider: React.FC<RoomProviderProps> = ({ children }) => {
  const [room, setRoom] = useState<Room | null>(null);
  const [roomState, setRoomStateInternal] = useState<RoomState | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [currentUser, setCurrentUser] = useState<User | null>(null);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const createRoom = useCallback(async (roomName: string, nickname: string): Promise<string> => {
    setIsLoading(true);
    setError(null);

    try {
      const request: CreateRoomRequest = {
        name: roomName,
        voting_system: 'fibonacci',
        auto_reveal: false,
      };

      const response = await api.createRoom(request);

      // Convert CreateRoomResponse to Room format
      const newRoom: Room = {
        id: response.id,
        name: response.name,
        voting_system: response.voting_system,
        auto_reveal: response.auto_reveal,
        created_at: response.created_at,
        updated_at: response.created_at,
      };

      setRoom(newRoom);

      // Generate a user ID for the creator
      const userId = crypto.randomUUID();
      setCurrentUserId(userId);
      setCurrentUser({
        id: userId,
        name: nickname,
        isVoted: false,
      });

      return response.id;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create room';
      setError(errorMessage);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const joinRoom = useCallback(async (roomId: string, nickname: string): Promise<void> => {
    setIsLoading(true);
    setError(null);

    try {
      // Fetch room details
      const roomData = await api.getRoom(roomId);
      setRoom(roomData);

      // Generate a user ID
      const userId = crypto.randomUUID();
      setCurrentUserId(userId);
      setCurrentUser({
        id: userId,
        name: nickname,
        isVoted: false,
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to join room';
      setError(errorMessage);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const leaveRoom = useCallback(async (): Promise<void> => {
    if (!room || !currentUserId) {
      return;
    }

    try {
      await api.leaveRoom(room.id, currentUserId);
      setRoom(null);
      setRoomStateInternal(null);
      setCurrentUserId(null);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to leave room';
      setError(errorMessage);
      throw err;
    }
  }, [room, currentUserId]);

  const setRoomState = useCallback((state: RoomState) => {
    setRoomStateInternal(state);
  }, []);

  const updateUsers = useCallback((users: User[]) => {
    setRoomStateInternal((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        users,
      };
    });
  }, []);

  const updateUserVoteStatus = useCallback((userId: string, hasVoted: boolean) => {
    setRoomStateInternal((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        users: prev.users.map(user =>
          user.id === userId
            ? { ...user, isVoted: hasVoted }
            : user
        ),
      };
    });
  }, []);

  const updateVotes = useCallback((votes: Vote[]) => {
    setRoomStateInternal((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        votes,
      };
    });
  }, []);

  const setRevealed = useCallback((revealed: boolean, average?: number | null) => {
    setRoomStateInternal((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        isRevealed: revealed,
        average: revealed ? average : undefined,
      };
    });
  }, []);

  const value: RoomContextState = useMemo(() => ({
    room,
    roomState,
    isLoading,
    error,
    currentUserId,
    currentUser,
    createRoom,
    joinRoom,
    leaveRoom,
    setRoomState,
    updateUsers,
    updateUserVoteStatus,
    updateVotes,
    setRevealed,
    clearError,
  }), [
    room,
    roomState,
    isLoading,
    error,
    currentUserId,
    currentUser,
    createRoom,
    joinRoom,
    leaveRoom,
    setRoomState,
    updateUsers,
    updateUserVoteStatus,
    updateVotes,
    setRevealed,
    clearError,
  ]);

  return <RoomContext.Provider value={value}>{children}</RoomContext.Provider>;
};

export const useRoom = (): RoomContextState => {
  const context = useContext(RoomContext);
  if (context === undefined) {
    throw new Error('useRoom must be used within a RoomProvider');
  }
  return context;
};
