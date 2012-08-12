#include <SDL/SDL.h>
#include <stdio.h>
#include <string.h>

typedef struct {
  int size;
  int pos;
  Uint8* data;
} buffer;

static buffer buf = {0, 0, NULL};

void stop()
{
  SDL_LockAudio();
  SDL_PauseAudio(1);
  free(buf.data);
  buf.size = buf.pos = 0;
  buf.data = NULL;
  SDL_UnlockAudio();
}

void audioCallback(void* userdata, uint8_t* stream, int len)
{
  int available = buf.size - buf.pos;
  int nbytes = available < len ? available : len;
  Uint8* ptr = &buf.data[buf.pos];
  int i;

  for (i = 0; i < len; i++)
    stream[i] = i < nbytes ? *ptr++ : 0;

  buf.pos += nbytes;

  // Out of sound, turn off the audio.
  if (buf.pos >= buf.size)
  {
    stop();
  }
}

int openAudio(SDL_AudioSpec* spec)
{
  int ret = 0;
  spec->callback = audioCallback;
  ret = SDL_OpenAudio(spec, NULL);
  if (ret != 0) {
    fprintf(stderr, "Error opening audio: %s\n", SDL_GetError());
  }
  return ret;
}

void play(Uint8* data, int nbytes)
{
  int newSize = nbytes;
  int notPlaying = 1;
  Uint8* dataPtr;

  SDL_LockAudio();

  // There's existing audio queued up, add it to the new buffer
  if (buf.data) {
    newSize += buf.size - buf.pos;
  }

  buffer newBuf;
  newBuf.data = (Uint8*)malloc(newSize);
  dataPtr = newBuf.data;
  newBuf.size = newSize;
  newBuf.pos = 0;

  if (buf.data) {
    // Copy old buffer to new, free old buffer.
    memcpy(dataPtr, &buf.data[buf.pos], buf.size - buf.pos);
    free(buf.data);
    dataPtr += buf.size - buf.pos;
  }

  memcpy(dataPtr, data, nbytes);

  buf = newBuf;

  SDL_UnlockAudio();

  SDL_PauseAudio(0);
}
