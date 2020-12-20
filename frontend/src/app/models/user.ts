export class User {
    id: number;
    username: string;
    firstName: string;
    lastName: string;
    age: number;
    gender: number;
    city: string;
    interests: string[];
    isFriend: boolean;
}

export class Session {
    token: string;
    userId: number;
    username: string;
}
