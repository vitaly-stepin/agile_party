// Room related types
export interface Room {
  id: string;
  name: string;
  voting_system: string;
  auto_reveal: boolean;
  created_at: string;
  updated_at: string;
}

export interface RoomSettings {
  voting_system: string;
  auto_reveal: boolean;
}

export interface NewRoomReq {
  name: string;
  voting_system?: string;
  auto_reveal?: boolean;
  settings?: RoomSettings;
}

export interface NewRoomResp {
  id: string;
  name: string;
  voting_system: string;
  auto_reveal: boolean;
  created_at: string;
}

// User related types
export interface User {
  id: string;
  userId?: string; // Legacy compatibility
  name: string;
  isVoted: boolean;
  hasVoted?: boolean; // Legacy compatibility
  isOnline?: boolean;
}

// Vote related types
export interface Vote {
  userId: string;
  userName: string;
  value: string;
}

// Room state types
export interface RoomState {
  roomId: string;
  roomName: string;
  users: User[];
  votes: Vote[];
  isRevealed: boolean;
  average?: number | null;
  taskDescription?: string;
}

// WebSocket event types
export type ClientEventType = 'vote' | 'reveal' | 'clear' | 'update_nickname' | 'set_task';
export type ServerEventType =
  | 'room_state'
  | 'user_joined'
  | 'user_left'
  | 'vote_submitted'
  | 'votes_revealed'
  | 'votes_cleared'
  | 'user_updated'
  | 'error';

export interface ClientEvent<T = any> {
  type: ClientEventType;
  payload: T;
}

export interface ServerEvent<T = any> {
  type: ServerEventType;
  payload: T;
}

// Client event payloads
export interface VotePayload {
  value: string;
}

export interface UpdateNicknamePayload {
  nickname: string;
}

export interface SetTaskPayload {
  description: string;
}

// Server event payloads
export interface RoomStatePayload {
  roomId: string;
  roomName: string;
  users: User[];
  votes: Vote[];
  isRevealed: boolean;
  average?: number | null;
}

export interface UserJoinedPayload {
  userId: string;
  name: string;
  isVoted: boolean;
  isOnline: boolean;
}

export interface UserLeftPayload {
  userId: string;
}

export interface VoteSubmittedPayload {
  userId: string;
  hasVoted: boolean;
}

export interface VotesRevealedPayload {
  votes: Vote[];
  average?: number | null;
}

export interface UserUpdatedPayload {
  userId: string;
  name: string;
}

export interface ErrorPayload {
  message: string;
  code?: string;
}

// Valid vote values for DBS Fibonacci
export const VALID_VOTES = ['0', '0.5', '1', '2', '3', '5', '8', '13', '20', '40', '100', '?'] as const;
export type ValidVote = typeof VALID_VOTES[number];
export type VoteValue = typeof VALID_VOTES[number];
