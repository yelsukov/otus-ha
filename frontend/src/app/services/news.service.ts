import {Injectable} from '@angular/core';
import {webSocket, WebSocketSubject} from 'rxjs/webSocket';

import {AuthService} from '@services/auth.service';
import {environment} from '@env/environment';

export const WS_ENDPOINT = environment.wsApiUrl + '/news';

@Injectable({
    providedIn: 'root'
})
export class NewsService {

    private socket$: WebSocketSubject<any>;

    constructor(private authService: AuthService) {
    }

    public connect(): any {
        if (this.isConnected()) {
            return this.socket$;
        }

        const session = this.authService.currentSessionValue;
        if (!session || !session.token) {
            throw new Error('unauthorized attempt to connect wit socket');
        }

        this.socket$ = webSocket({
            url: WS_ENDPOINT + '?token=' + session.token,
            openObserver: {
                next: () => console.log('[NewsService]: connection opened'),
                error: error => console.log(error)
            },
            closeObserver: {
                next: () => console.log('[NewsService]: connection closed')
            },
        });

        return this.socket$;
    }

    public isConnected() {
        return !!this.socket$ && !this.socket$.closed;
    }

    send(msg: any) {
        this.socket$.next(msg);
    }

    close() {
        this.socket$.complete();
        delete this.socket$;
    }
}
