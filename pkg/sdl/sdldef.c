#include <SDL/SDL.h>

/* Template for converting SDL types and constants into Go.
 * After banthar's Go-SDL bindings (http://github.com/banthar/Go-SDL). */

typedef SDL_Surface $surface;
typedef SDL_PixelFormat $pixelFormat;
/* Omitting SDL_Rect, since it's field names collide with the IntRect
 * interface used. Writing it by hand instead. */
typedef SDL_Color $color;
typedef SDL_Palette $palette;
typedef SDL_VideoInfo $videoInfo;
typedef SDL_Overlay $overlay;

typedef SDL_ActiveEvent $activeEvent;
typedef SDL_KeyboardEvent $keyboardEvent;
typedef SDL_MouseMotionEvent $mouseMotionEvent;
typedef SDL_MouseButtonEvent $mouseButtonEvent;
typedef SDL_JoyAxisEvent $joyAxisEvent;
typedef SDL_JoyBallEvent $joyBallEvent;
typedef SDL_JoyHatEvent $joyHatEvent;
typedef SDL_JoyButtonEvent $joyButtonEvent;
typedef SDL_ResizeEvent $resizeEvent;
typedef SDL_ExposeEvent $exposeEvent;
typedef SDL_QuitEvent $quitEvent;
typedef SDL_UserEvent $userEvent;
typedef SDL_SysWMmsg $sysWMmsg;
typedef SDL_SysWMEvent $sysWMEvent;
typedef SDL_Event $event;
typedef SDL_keysym $keysym;

enum {
  // Byte order
  $BYTEORDER = SDL_BYTEORDER,
  $BIG_ENDIAN = SDL_BIG_ENDIAN,
  $LIL_ENDIAN = SDL_LIL_ENDIAN,

  // Init flags
  $INIT_AUDIO = SDL_INIT_AUDIO,
  $INIT_VIDEO = SDL_INIT_VIDEO,

  // Video flags
  $SWSURFACE = SDL_SWSURFACE,
  $HWSURFACE = SDL_HWSURFACE,
  $DOUBLEBUF = SDL_DOUBLEBUF,
  $FULLSCREEN = SDL_FULLSCREEN,
  $RESIZABLE = SDL_RESIZABLE,
  $RLEACCEL = SDL_RLEACCEL,
  $ASYNCBLIT = SDL_ASYNCBLIT,

  // Event types
  $NOEVENT = SDL_NOEVENT,
  $ACTIVEEVENT = SDL_ACTIVEEVENT,
  $KEYDOWN = SDL_KEYDOWN,
  $KEYUP = SDL_KEYUP,
  $MOUSEMOTION = SDL_MOUSEMOTION,
  $MOUSEBUTTONDOWN = SDL_MOUSEBUTTONDOWN,
  $MOUSEBUTTONUP = SDL_MOUSEBUTTONUP,
  $JOYAXISMOTION = SDL_JOYAXISMOTION,
  $JOYBALLMOTION = SDL_JOYBALLMOTION,
  $JOYHATMOTION = SDL_JOYHATMOTION,
  $JOYBUTTONDOWN = SDL_JOYBUTTONDOWN,
  $JOYBUTTONUP = SDL_JOYBUTTONUP,
  $QUIT = SDL_QUIT,
  $SYSWMEVENT = SDL_SYSWMEVENT,
  $VIDEORESIZE = SDL_VIDEORESIZE,
  $VIDEOEXPOSE = SDL_VIDEOEXPOSE,
  $USEREVENT = SDL_USEREVENT,
  $NUMEVENTS = SDL_NUMEVENTS,

  // Event masks
  $ACTIVEEVENTMASK = SDL_ACTIVEEVENTMASK,
  $KEYDOWNMASK = SDL_KEYDOWNMASK,
  $KEYUPMASK = SDL_KEYUPMASK,
  $KEYEVENTMASK = SDL_KEYEVENTMASK,
  $MOUSEMOTIONMASK = SDL_MOUSEMOTIONMASK,
  $MOUSEBUTTONDOWNMASK = SDL_MOUSEBUTTONDOWNMASK,
  $MOUSEBUTTONUPMASK = SDL_MOUSEBUTTONUPMASK,
  $MOUSEEVENTMASK = SDL_MOUSEEVENTMASK,
  $JOYAXISMOTIONMASK = SDL_JOYAXISMOTIONMASK,
  $JOYBALLMOTIONMASK = SDL_JOYBALLMOTIONMASK,
  $JOYHATMOTIONMASK = SDL_JOYHATMOTIONMASK,
  $JOYBUTTONDOWNMASK = SDL_JOYBUTTONDOWNMASK,
  $JOYBUTTONUPMASK = SDL_JOYBUTTONUPMASK,
  $JOYEVENTMASK = SDL_JOYEVENTMASK,
  $VIDEORESIZEMASK = SDL_VIDEORESIZEMASK,
  $VIDEOEXPOSEMASK = SDL_VIDEOEXPOSEMASK,
  $QUITMASK = SDL_QUITMASK,
  $SYSWMEVENTMASK = SDL_SYSWMEVENTMASK,

  // Keys
  $K_UNKNOWN = SDLK_UNKNOWN,
  $K_FIRST = SDLK_FIRST,
  $K_BACKSPACE = SDLK_BACKSPACE,
  $K_TAB = SDLK_TAB,
  $K_CLEAR = SDLK_CLEAR,
  $K_RETURN = SDLK_RETURN,
  $K_PAUSE = SDLK_PAUSE,
  $K_ESCAPE = SDLK_ESCAPE,
  $K_SPACE = SDLK_SPACE,
  $K_EXCLAIM = SDLK_EXCLAIM,
  $K_QUOTEDBL = SDLK_QUOTEDBL,
  $K_HASH = SDLK_HASH,
  $K_DOLLAR = SDLK_DOLLAR,
  $K_AMPERSAND = SDLK_AMPERSAND,
  $K_QUOTE = SDLK_QUOTE,
  $K_LEFTPAREN = SDLK_LEFTPAREN,
  $K_RIGHTPAREN = SDLK_RIGHTPAREN,
  $K_ASTERISK = SDLK_ASTERISK,
  $K_PLUS = SDLK_PLUS,
  $K_COMMA = SDLK_COMMA,
  $K_MINUS = SDLK_MINUS,
  $K_PERIOD = SDLK_PERIOD,
  $K_SLASH = SDLK_SLASH,
  $K_0 = SDLK_0,
  $K_1 = SDLK_1,
  $K_2 = SDLK_2,
  $K_3 = SDLK_3,
  $K_4 = SDLK_4,
  $K_5 = SDLK_5,
  $K_6 = SDLK_6,
  $K_7 = SDLK_7,
  $K_8 = SDLK_8,
  $K_9 = SDLK_9,
  $K_COLON = SDLK_COLON,
  $K_SEMICOLON = SDLK_SEMICOLON,
  $K_LESS = SDLK_LESS,
  $K_EQUALS = SDLK_EQUALS,
  $K_GREATER = SDLK_GREATER,
  $K_QUESTION = SDLK_QUESTION,
  $K_AT = SDLK_AT,
  $K_LEFTBRACKET = SDLK_LEFTBRACKET,
  $K_BACKSLASH = SDLK_BACKSLASH,
  $K_RIGHTBRACKET = SDLK_RIGHTBRACKET,
  $K_CARET = SDLK_CARET,
  $K_UNDERSCORE = SDLK_UNDERSCORE,
  $K_BACKQUOTE = SDLK_BACKQUOTE,
  $K_a = SDLK_a,
  $K_b = SDLK_b,
  $K_c = SDLK_c,
  $K_d = SDLK_d,
  $K_e = SDLK_e,
  $K_f = SDLK_f,
  $K_g = SDLK_g,
  $K_h = SDLK_h,
  $K_i = SDLK_i,
  $K_j = SDLK_j,
  $K_k = SDLK_k,
  $K_l = SDLK_l,
  $K_m = SDLK_m,
  $K_n = SDLK_n,
  $K_o = SDLK_o,
  $K_p = SDLK_p,
  $K_q = SDLK_q,
  $K_r = SDLK_r,
  $K_s = SDLK_s,
  $K_t = SDLK_t,
  $K_u = SDLK_u,
  $K_v = SDLK_v,
  $K_w = SDLK_w,
  $K_x = SDLK_x,
  $K_y = SDLK_y,
  $K_z = SDLK_z,
  $K_DELETE = SDLK_DELETE,
  $K_KP0 = SDLK_KP0,
  $K_KP1 = SDLK_KP1,
  $K_KP2 = SDLK_KP2,
  $K_KP3 = SDLK_KP3,
  $K_KP4 = SDLK_KP4,
  $K_KP5 = SDLK_KP5,
  $K_KP6 = SDLK_KP6,
  $K_KP7 = SDLK_KP7,
  $K_KP8 = SDLK_KP8,
  $K_KP9 = SDLK_KP9,
  $K_KP_PERIOD = SDLK_KP_PERIOD,
  $K_KP_DIVIDE = SDLK_KP_DIVIDE,
  $K_KP_MULTIPLY = SDLK_KP_MULTIPLY,
  $K_KP_MINUS = SDLK_KP_MINUS,
  $K_KP_PLUS = SDLK_KP_PLUS,
  $K_KP_ENTER = SDLK_KP_ENTER,
  $K_KP_EQUALS = SDLK_KP_EQUALS,
  $K_UP = SDLK_UP,
  $K_DOWN = SDLK_DOWN,
  $K_RIGHT = SDLK_RIGHT,
  $K_LEFT = SDLK_LEFT,
  $K_INSERT = SDLK_INSERT,
  $K_HOME = SDLK_HOME,
  $K_END = SDLK_END,
  $K_PAGEUP = SDLK_PAGEUP,
  $K_PAGEDOWN = SDLK_PAGEDOWN,
  $K_F1 = SDLK_F1,
  $K_F2 = SDLK_F2,
  $K_F3 = SDLK_F3,
  $K_F4 = SDLK_F4,
  $K_F5 = SDLK_F5,
  $K_F6 = SDLK_F6,
  $K_F7 = SDLK_F7,
  $K_F8 = SDLK_F8,
  $K_F9 = SDLK_F9,
  $K_F10 = SDLK_F10,
  $K_F11 = SDLK_F11,
  $K_F12 = SDLK_F12,
  $K_F13 = SDLK_F13,
  $K_F14 = SDLK_F14,
  $K_F15 = SDLK_F15,
  $K_NUMLOCK = SDLK_NUMLOCK,
  $K_CAPSLOCK = SDLK_CAPSLOCK,
  $K_SCROLLOCK = SDLK_SCROLLOCK,
  $K_RSHIFT = SDLK_RSHIFT,
  $K_LSHIFT = SDLK_LSHIFT,
  $K_RCTRL = SDLK_RCTRL,
  $K_LCTRL = SDLK_LCTRL,
  $K_RALT = SDLK_RALT,
  $K_LALT = SDLK_LALT,
  $K_RMETA = SDLK_RMETA,
  $K_LMETA = SDLK_LMETA,
  $K_LSUPER = SDLK_LSUPER,
  $K_RSUPER = SDLK_RSUPER,
  $K_MODE = SDLK_MODE,
  $K_COMPOSE = SDLK_COMPOSE,
  $K_HELP = SDLK_HELP,
  $K_PRINT = SDLK_PRINT,
  $K_SYSREQ = SDLK_SYSREQ,
  $K_BREAK = SDLK_BREAK,
  $K_MENU = SDLK_MENU,
  $K_POWER = SDLK_POWER,
  $K_EURO = SDLK_EURO,
  $K_UNDO = SDLK_UNDO,

  // Key mods

  $KMOD_NONE = KMOD_NONE,
  $KMOD_LSHIFT = KMOD_LSHIFT,
  $KMOD_RSHIFT = KMOD_RSHIFT,
  $KMOD_LCTRL = KMOD_LCTRL,
  $KMOD_RCTRL = KMOD_RCTRL,
  $KMOD_LALT = KMOD_LALT,
  $KMOD_RALT = KMOD_RALT,
  $KMOD_LMETA = KMOD_LMETA,
  $KMOD_RMETA = KMOD_RMETA,
  $KMOD_NUM = KMOD_NUM,
  $KMOD_CAPS = KMOD_CAPS,
  $KMOD_MODE = KMOD_MODE,

};
