import type { CreateRoomRequest, CreateRoomResponse, Room, RoomState } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = 'ApiError';
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorText = await response.text();
    let errorMessage = `HTTP ${response.status}: ${response.statusText}`;

    try {
      const errorJson = JSON.parse(errorText);
      errorMessage = errorJson.error || errorJson.message || errorMessage;
    } catch {
      // If response is not JSON, use the text
      errorMessage = errorText || errorMessage;
    }

    throw new ApiError(response.status, errorMessage);
  }

  return response.json();
}

export const api = {
  /**
   * Create a new room
   */
  async createRoom(request: CreateRoomRequest): Promise<CreateRoomResponse> {
    const response = await fetch(`${API_BASE_URL}/api/rooms`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    return handleResponse<CreateRoomResponse>(response);
  },

  /**
   * Get room details by ID
   */
  async getRoom(roomId: string): Promise<Room> {
    const response = await fetch(`${API_BASE_URL}/api/rooms/${roomId}`);
    return handleResponse<Room>(response);
  },

  /**
   * Get current room state (live state)
   */
  async getRoomState(roomId: string): Promise<RoomState> {
    const response = await fetch(`${API_BASE_URL}/api/rooms/${roomId}/state`);
    return handleResponse<RoomState>(response);
  },

  /**
   * Join a room
   */
  async joinRoom(roomId: string, userId: string, userName: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/rooms/${roomId}/users`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ user_id: userId, user_name: userName }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(response.status, errorText);
    }
  },

  /**
   * Leave a room
   */
  async leaveRoom(roomId: string, userId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/rooms/${roomId}/users/${userId}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(response.status, errorText);
    }
  },

  /**
   * Update user name
   */
  async updateUserName(roomId: string, userId: string, newName: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/rooms/${roomId}/users/${userId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ new_name: newName }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(response.status, errorText);
    }
  },

  /**
   * Submit a vote
   */
  async submitVote(roomId: string, userId: string, value: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/rooms/${roomId}/votes`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ user_id: userId, value }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(response.status, errorText);
    }
  },

  /**
   * Reveal votes
   */
  async revealVotes(roomId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/rooms/${roomId}/reveal`, {
      method: 'POST',
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(response.status, errorText);
    }
  },

  /**
   * Clear votes (start new round)
   */
  async clearVotes(roomId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/rooms/${roomId}/clear`, {
      method: 'POST',
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(response.status, errorText);
    }
  },

  /**
   * Health check
   */
  async healthCheck(): Promise<{ status: string }> {
    const response = await fetch(`${API_BASE_URL}/api/health`);
    return handleResponse<{ status: string }>(response);
  },
};

export { ApiError };
